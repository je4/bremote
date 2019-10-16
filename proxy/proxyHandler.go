package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/je4/bremote/api"
	"github.com/je4/bremote/common"
	"github.com/mintance/go-uniqid"
	"github.com/op/go-logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

type ProxyServiceServer struct {
	log          *logging.Logger
	proxySession *ProxySession
}

func NewProxyServiceServer(proxySession *ProxySession, log *logging.Logger) *ProxyServiceServer {
	pss := &ProxyServiceServer{proxySession: proxySession, log: log}
	return pss
}

func (pss ProxyServiceServer) Ping(ctx context.Context, param *pb.String) (*pb.String, error) {
	traceId, sourceInstance, err := common.RpcContextMetadata2(ctx)
	if err != nil {
		pss.log.Errorf("invalid metadata in call to %v: %v", "StartBrowser()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	pss.log.Infof("[%v] %v -> /Ping( %v )", traceId, sourceInstance, param.GetValue())

	ret := new(pb.String)
	ret.Value = "pong: " + param.GetValue()

	return ret, nil
}

func (pss ProxyServiceServer) Init(ctx context.Context, param *pb.InitParam) (*empty.Empty, error) {
	traceId, sourceInstance, err := common.RpcContextMetadata2(ctx)
	if err != nil {
		pss.log.Errorf("invalid metadata in call to %v: %v", "StartBrowser()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}
	client := param.GetInstance()
	stat := param.GetStatus()
	httpAddr := param.GetHttpAddr()

	pss.log.Infof("[%v] %v -> /Init( %v, %v, %v )", traceId, sourceInstance, client, pb.ProxySessionType_name[int32(param.GetSessionType())], stat)

	instance := pss.proxySession.GetInstance()
	// set session typ only, if master key or undefined
	currentType := pss.proxySession.GetSessionType()
	newType := common.SessionType(param.SessionType)
	if currentType != newType {
		if currentType == common.SessionType_Undefined || pss.proxySession.IsGeneric() {
			pss.proxySession.SetSessionType(newType)
		} else {
			pss.log.Errorf("forbidden to change sessionType from %v to %v", currentType, newType)
			return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("forbidden to change sessionType from %v to %v", currentType, newType))
		}
	}

	if client != instance {
		if pss.proxySession.IsGeneric() {
			if _, err := pss.proxySession.GetProxy().GetSession(client); err == nil {
				return nil, status.Errorf(codes.AlreadyExists, fmt.Sprintf("cannot rename %v to %v: instance already exists", instance, client, err))
			}
			if err := pss.proxySession.SetInstance(client); err != nil {
				return nil, status.Errorf(codes.OutOfRange, fmt.Sprintf("cannot rename %v to %v: %v", instance, client, err))
			}
		} else {
			pss.log.Errorf("forbidden to rename %v to %v", instance, client)
			return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("forbidden to rename %v to %v", instance, client))
		}
	}
	go func() {
		time.Sleep(time.Millisecond * 300)

		for name, session := range pss.proxySession.GetProxy().GetSessions() {
			// don't notify myself
			if name == client {
				continue
			}
			// send message to all controllers
			if session.GetSessionType() != common.SessionType_Controller {
				continue
			}
			cw := pb.NewControllerWrapper(name, session.GetSessionPtr())

			traceId = uniqid.New(uniqid.Params{"traceid_", false})
			pss.log.Infof("[%v] announcing %v to %v", traceId, client, name)
			if err := cw.NewClient(traceId, name, client, stat, httpAddr); err != nil {
				pss.log.Errorf("[%v] error announcing %v to %v: %v", traceId, client, name, err)
			}
			// send only to first controller...
			break;
		}
	}()

	return new(empty.Empty), nil
}

func (pss ProxyServiceServer) GetClients(ctx context.Context, req *pb.GetClientsParam) (*pb.ProxyClientList, error) {
	clients := new(pb.ProxyClientList)
	clients.Clients = []*pb.ProxyClient{}
	withStatus := req.GetWithStatus()
	for name, session := range pss.proxySession.GetProxy().GetSessions() {
		// ignore yourself
		if name == pss.proxySession.GetInstance() {
			continue
		}
		status := common.ClientStatus_Empty
		if withStatus {
			cw := pb.NewClientWrapper(pss.proxySession.GetInstance(), session.GetSessionPtr())
			traceId := uniqid.New(uniqid.Params{"traceid_", false})
			pss.log.Debugf("[%v] getting status of %v", traceId, name)
			s, err := cw.GetStatus(traceId, name)
			if err != nil {
				pss.log.Errorf("[%v] error getting status of %v: %v", traceId, name, err)
			}
			status = s
		}
		clients.Clients = append(clients.Clients, &pb.ProxyClient{
			Instance: session.GetInstance(),
			Type:     pb.ProxySessionType(session.GetSessionType()),
			Status:   status,
		})
	}
	return clients, nil
}

func (pss ProxyServiceServer) GroupAddInstance(ctx context.Context, req *pb.GroupInstanceMessage) (*empty.Empty, error) {
	groupName := req.GetGroup()
	instanceName := req.GetInstance()
	pss.proxySession.proxy.groups.AddInstance(groupName, instanceName)
	return &empty.Empty{}, nil
}
func (pss ProxyServiceServer) GroupRemoveInstance(ctx context.Context, req *pb.GroupInstanceMessage) (*empty.Empty, error) {
	groupName := req.GetGroup()
	instanceName := req.GetInstance()
	pss.proxySession.proxy.groups.RemoveInstance(groupName, instanceName)
	return &empty.Empty{}, nil
}
func (pss ProxyServiceServer) GroupGetMembers(ctx context.Context, req *pb.String) (*pb.MemberListMessage, error) {
	groupName := req.GetValue()

	clients := new(pb.MemberListMessage)
	clients.Instances = pss.proxySession.proxy.groups.GetMembers(groupName)
	return clients, nil
}
func (pss ProxyServiceServer) GroupDelete(ctx context.Context, req *pb.String) (*empty.Empty, error) {
	pss.proxySession.proxy.groups.Delete(req.Value)
	return &empty.Empty{}, nil
}

func (pss ProxyServiceServer) GroupList(context.Context, *empty.Empty) (*pb.GroupListMessage, error) {
	result := &pb.GroupListMessage{}
	result.Groups = pss.proxySession.proxy.groups.GetGroups()
	return result, nil
}

func (pss ProxyServiceServer) WebsocketMessage(ctx context.Context, req *pb.Bytes) (*empty.Empty, error) {
	traceId, sourceInstance, targetGroup, err := common.RpcContextMetadata(ctx)
	if err != nil {
		pss.log.Errorf("invalid metadata in call to %v: %v", "WebSocketMessage()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata in call to %v: %v", "WebSocketMessage()", err))
	}

	pss.log.Infof("[%v] %v -> /ws() -> %v", traceId, sourceInstance, targetGroup)

	instances := pss.proxySession.proxy.groups.GetMembers(targetGroup)
	for _, instanceName := range instances {
		// don't send to myself
		if instanceName == pss.proxySession.GetInstance() {
			continue
		}
		session, err := pss.proxySession.proxy.GetSession(instanceName)
		if err != nil {
			pss.log.Errorf("client %v not active - cannot send websocket message: %v", instanceName, err)
			continue
		}
		// every session has it's own grpc service
		if session.GetSessionType() == common.SessionType_Client {
			cw := pb.NewClientWrapper(pss.proxySession.GetInstance(), session.GetSessionPtr())
			pss.log.Debugf("[%v] sending websocket message to %v", traceId, session.GetInstance())
			err = cw.WebsocketMessage(traceId, sourceInstance, targetGroup, req.GetValue())
			if err != nil {
				pss.log.Errorf("[%v] sending websocket message to %v: %v", traceId, session.GetInstance(), err)
			}
		}
	}

	return &empty.Empty{}, nil
}

func (pss ProxyServiceServer) KVStoreSetValue(ctx context.Context, req *pb.KVSetValueMessage) (*empty.Empty, error) {
	traceId, sourceInstance, err := common.RpcContextMetadata2(ctx)
	if err != nil {
		pss.log.Errorf("invalid metadata in call to %v: %v", "KVStoreSetValue()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata in call to %v: %v", "KVStoreSetValue()", err))
	}

	k := req.GetKey()
	key := fmt.Sprintf("%s-%s", k.GetClient(), k.GetKey())

	pss.log.Infof("[%v] %v -> /KVStoreSetValue(%s) -> %v", traceId, sourceInstance, key)

	if err := pss.proxySession.proxy.setVar(key, req.GetValue()); err != nil {
		pss.log.Errorf("cannot set data for %v from key value store: %v", key, err)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("cannot set data for %v from key value store: %v", key, err))
	}
	return &empty.Empty{}, nil
}

func (pss ProxyServiceServer) KVStoreGetValue(ctx context.Context, req *pb.KVKeyMessage) (*pb.String, error) {
	traceId, sourceInstance, err := common.RpcContextMetadata2(ctx)
	if err != nil {
		pss.log.Errorf("invalid metadata in call to %v: %v", "KVStoreGetValue()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata in call to %v: %v", "KVStoreGetValue()", err))
	}

	key := fmt.Sprintf("%s-%s", req.GetClient(), req.GetKey())

	pss.log.Infof("[%v] %v -> /KVStoreGetValue(%s)", traceId, sourceInstance, key)

	ret, err := pss.proxySession.proxy.getVar(key)
	if err != nil {
		pss.log.Errorf("cannot get data for %v from key value store: %v", key, err)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("cannot get data for %v from key value store: %v", key, err))
	}
	return &pb.String{Value: ret}, nil
}
func (pss ProxyServiceServer) KVStoreDeleteValue(ctx context.Context, req *pb.KVKeyMessage) (*empty.Empty, error) {
	traceId, sourceInstance, err := common.RpcContextMetadata2(ctx)
	if err != nil {
		pss.log.Errorf("invalid metadata in call to %v: %v", "KVStoreDeleteValue()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata in call to %v: %v", "KVStoreDeleteValue()", err))
	}

	key := fmt.Sprintf("%s-%s", req.GetClient(), req.GetKey())

	pss.log.Infof("[%v] %v -> /KVStoreDeleteValue(%s) -> %v", traceId, sourceInstance, key)

	if err := pss.proxySession.proxy.deleteVar(key); err != nil {
		pss.log.Errorf("cannot delete data for %v from key value store: %v", key, err)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("cannot delete data for %v from key value store: %v", key, err))
	}

	return &empty.Empty{}, nil
}

func (pss ProxyServiceServer) KVStoreList(ctx context.Context, req *empty.Empty) (*pb.KVSetValueListMessage, error) {
	traceId, sourceInstance, err := common.RpcContextMetadata2(ctx)
	if err != nil {
		pss.log.Errorf("invalid metadata in call to %v: %v", "KVStoreList()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata in call to %v: %v", "KVStoreList()", err))
	}

	pss.log.Infof("[%v] %v -> /KVStoreList() -> %v", traceId, sourceInstance)

	keys, err := pss.proxySession.proxy.getKeys("")
	if err != nil {
		pss.log.Errorf("cannot get keys from key value store: %v", err)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("cannot get keys from key value store: %v", err))
	}
	result := &pb.KVSetValueListMessage{}
	for _, key := range keys {
		value, err := pss.proxySession.proxy.getVar(key)
		if err != nil {
			pss.log.Errorf("cannot get data for %v from key value store: %v", key, err)
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("cannot get data for %v from key value store: %v", key, err))
		}
		kparts := strings.Split(key, "-")
		if len(kparts) < 2 {
			pss.log.Errorf("invalid key %v", key)
			continue
		}
		client, key := kparts[0], kparts[1:]
		result.Data = append(result.Data, &pb.KVSetValueMessage{
			Key: &pb.KVKeyMessage{
				Client: client,
				Key:    strings.Join(key, "-"),
			},
			Value: value,
		})
	}

	return result, nil
}
func (pss ProxyServiceServer) KVStoreClientList(ctx context.Context, req *pb.String) (*pb.KVSetValueListMessage, error) {
	traceId, sourceInstance, err := common.RpcContextMetadata2(ctx)
	if err != nil {
		pss.log.Errorf("invalid metadata in call to %v: %v", "KVStoreClientList()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata in call to %v: %v", "KVStoreClientList()", err))
	}

	pss.log.Infof("[%v] %v -> /KVStoreClientList() -> %v", traceId, sourceInstance)

	keys, err := pss.proxySession.proxy.getKeys(req.GetValue() + "-")
	if err != nil {
		pss.log.Errorf("cannot get keys from key value store: %v", err)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("cannot get keys from key value store: %v", err))
	}
	result := &pb.KVSetValueListMessage{}
	if len(keys) > 0 {
		for _, key := range keys {
			value, err := pss.proxySession.proxy.getVar(key)
			if err != nil {
				pss.log.Errorf("cannot get data for %v from key value store: %v", key, err)
				return nil, status.Errorf(codes.NotFound, fmt.Sprintf("cannot get data for %v from key value store: %v", key, err))
			}
			kparts := strings.Split(key, "-")
			if len(kparts) < 2 {
				pss.log.Errorf("invalid key %v", key)
				continue
			}
			client, key := kparts[0], kparts[1:]
			result.Data = append(result.Data, &pb.KVSetValueMessage{
				Key: &pb.KVKeyMessage{
					Client: client,
					Key:    strings.Join(key, "-"),
				},
				Value: value,
			})
		}
	}
	return result, nil
}
