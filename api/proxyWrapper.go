package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/je4/bremote/common"
	"github.com/mintance/go-uniqid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"net"
)

type ProxyWrapper struct {
	instanceName       string
	session            **yamux.Session
	proxyServiceClient *ProxyServiceClient
	conn               *grpc.ClientConn
}

func NewProxyWrapper(instanceName string, session **yamux.Session) *ProxyWrapper {
	cw := &ProxyWrapper{instanceName: instanceName, session: session, proxyServiceClient: nil}
	return cw
}

func (pw *ProxyWrapper) connect() (err error) {
	if *pw.session == nil {
		pw.proxyServiceClient = nil
		return errors.New(fmt.Sprintf("session closed"))
	}

	// it's a singleton
	if pw.proxyServiceClient != nil {
		return nil
	}
	// gRPC dial over incoming net.Conn
	// singleton!!!
	doDial := pw.conn == nil
	if pw.conn != nil {
		if pw.conn.GetState() == connectivity.TransientFailure {
			pw.conn.Close()
			doDial = true
		}
		if pw.conn.GetState() == connectivity.Shutdown {
			doDial = true
		}
	}
	if doDial {
		pw.conn, err = grpc.Dial(":7777", grpc.WithInsecure(),
			grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
				if *pw.session == nil {
					return nil, errors.New(fmt.Sprintf("session %s closed", s))
				}
				return (*pw.session).Open()
			}),
		)
		if err != nil {
			return errors.New("cannot dial grpc connection to :7777")
		}
	}
	c := NewProxyServiceClient(pw.conn)
	pw.proxyServiceClient = &c
	return nil
}

func (cw *ProxyWrapper) Ping(traceId string) (string, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}

	if err := cw.connect(); err != nil {
		return "", emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	pingResult, err := (*cw.proxyServiceClient).Ping(ctx, &String{Value: "ping"})
	if err != nil {
		return "", emperror.Wrapf(err, "error pinging")
	}
	return pingResult.GetValue(), nil
}

func (cw *ProxyWrapper) Init(traceId string, instance string, sessionType common.SessionType, status string, httpAddr string) error {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}

	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	_, err := (*cw.proxyServiceClient).Init(ctx, &InitParam{
		Instance:    instance,
		SessionType: ProxySessionType(sessionType),
		Status:      status,
		HttpAddr:    httpAddr,
	})
	if err != nil {
		return emperror.Wrapf(err, "error initializing instance")
	}
	return nil
}

func (cw *ProxyWrapper) GetClients(traceId string, t common.SessionType, withStatus bool) ([]common.ClientInfo, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return []common.ClientInfo{}, emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	clients, err := (*cw.proxyServiceClient).GetClients(ctx, &GetClientsParam{WithStatus: withStatus})
	if err != nil {
		return []common.ClientInfo{}, emperror.Wrap(err, "cannot get clients")
	}
	ret := []common.ClientInfo{}
	for _, c := range clients.GetClients() {
		// we only want we need
		if t != common.SessionType_All {
			if c.GetType() != ProxySessionType(t) {
				continue
			}
		}
		ret = append(ret, common.ClientInfo{
			InstanceName: c.GetInstance(),
			Type:         common.SessionType(c.GetType()),
			Status:       c.GetStatus(),
		})
	}
	return ret, nil
}

func (cw *ProxyWrapper) GroupAddInstance(traceId string, group string, instance string) error {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	_, err := (*cw.proxyServiceClient).GroupAddInstance(ctx, &GroupInstanceMessage{Group: group, Instance: instance})
	if err != nil {
		return emperror.Wrapf(err, "error adding instance %v to group &v", instance, group)
	}
	return nil
}

func (cw *ProxyWrapper) GroupRemoveInstance(traceId string, group string, instance string) error {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	_, err := (*cw.proxyServiceClient).GroupRemoveInstance(ctx, &GroupInstanceMessage{Group: group, Instance: instance})
	if err != nil {
		return emperror.Wrapf(err, "error adding instance %v to group &v", instance, group)
	}
	return nil
}

func (cw *ProxyWrapper) GroupGetMembers(traceId string, groupname string) ([]string, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return []string{}, emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	ret, err := (*cw.proxyServiceClient).GroupGetMembers(ctx, &String{Value: groupname})
	if err != nil {
		return []string{}, emperror.Wrapf(err, "error getting members of group %v", groupname)
	}
	return ret.GetInstances(), nil
}

func (cw *ProxyWrapper) GroupDelete(traceId string, groupname string) error {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	_, err := (*cw.proxyServiceClient).GroupDelete(ctx, &String{Value: groupname})
	if err != nil {
		return emperror.Wrapf(err, "error deleting group %v", groupname)
	}
	return nil
}

func (cw *ProxyWrapper) GroupList(traceId string) ([]string, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return []string{}, emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	ret, err := (*cw.proxyServiceClient).GroupList(ctx, &empty.Empty{})
	if err != nil {
		return []string{}, emperror.Wrapf(err, "error getting list of groups")
	}
	return ret.GetGroups(), nil
}

func (cw *ProxyWrapper) WebsocketMessage(traceId string, targetGroup string, data []byte) error {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(
		context.Background(),
		"sourceInstance", cw.instanceName,
		"targetInstance", targetGroup,
		"traceid", traceId)
	_, err := (*cw.proxyServiceClient).WebsocketMessage(ctx, &Bytes{Value: data})
	if err != nil {
		return emperror.Wrapf(err, "error getting list of groups")
	}
	return nil

}

func (cw *ProxyWrapper) KVStoreSetValue(client string, key string, value string, traceId string) error {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	_, err := (*cw.proxyServiceClient).KVStoreSetValue(ctx, &KVSetValueMessage{
		Key: &KVKeyMessage{
			Client: client,
			Key:    key,
		},
		Value: value,
	})
	if err != nil {
		return emperror.Wrapf(err, "[%v] error setting %s-%s to %s", traceId, client, key, value)
	}
	return nil
}

func (cw *ProxyWrapper) KVStoreGetValue(client string, key string, traceId string) (string, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return "", emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	ret, err := (*cw.proxyServiceClient).KVStoreGetValue(ctx, &KVKeyMessage{
		Client: client,
		Key:    key,
	})
	if err != nil {
		return "", emperror.Wrapf(err, "[%v] error getting %s-%s", traceId, client, key)
	}
	return ret.GetValue(), nil
}

func (cw *ProxyWrapper) KVStoreDeleteValue(client string, key string, traceId string) error {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	_, err := (*cw.proxyServiceClient).KVStoreDeleteValue(ctx, &KVKeyMessage{
		Client: client,
		Key:    key,
	})
	if err != nil {
		return emperror.Wrapf(err, "[%v] error getting %s-%s", traceId, client, key)
	}
	return nil
}

func (cw *ProxyWrapper) KVStoreList(traceId string) (*map[string]string, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return &map[string]string{}, emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	ret, err := (*cw.proxyServiceClient).KVStoreList(ctx, &empty.Empty{})
	if err != nil {
		return &map[string]string{}, emperror.Wrapf(err, "[%v] error getting kv store list", traceId)
	}
	result := map[string]string{}
	for _, val := range ret.GetData() {
		k := val.GetKey()
		key := fmt.Sprintf("%s-%s", k.GetClient(), k.GetKey())
		result[key] = val.GetValue()
	}
	return &result, nil
}

func (cw *ProxyWrapper) KVStoreClientList(client string, traceId string) (*map[string]string, error) {
	if traceId == "" {
		traceId = uniqid.New(uniqid.Params{"traceid_", false})
	}
	if err := cw.connect(); err != nil {
		return &map[string]string{}, emperror.Wrapf(err, "cannot connect")
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "sourceInstance", cw.instanceName, "traceid", traceId)
	ret, err := (*cw.proxyServiceClient).KVStoreClientList(ctx, &String{Value: client})
	if err != nil {
		return &map[string]string{}, emperror.Wrapf(err, "[%v] error getting kv store list", traceId)
	}
	result := &map[string]string{}
	for _, val := range ret.GetData() {
		k := val.GetKey()
		(*result)[k.GetKey()] = val.GetValue()
	}
	return result, nil
}
