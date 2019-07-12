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
	traceId, sourceInstance,  err := common.RpcContextMetadata2(ctx)
	if err != nil {
		pss.log.Errorf("invalid metadata in call to %v: %v", "StartBrowser()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}

	pss.log.Infof("[%v] %v -> /Ping( %v )", traceId, sourceInstance,  param.GetValue())

	ret := new(pb.String)
	ret.Value = "pong: " + param.GetValue()

	return ret, nil
}

func (pss ProxyServiceServer) Init(ctx context.Context, param *pb.InitParam) (*empty.Empty, error) {
	traceId, sourceInstance,  err := common.RpcContextMetadata2(ctx)
	if err != nil {
		pss.log.Errorf("invalid metadata in call to %v: %v", "StartBrowser()", err)
		return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("invalid metadata: %v", err))
	}
	client := param.GetInstance()
	stat := param.GetStatus()
	httpAddr := param.GetHttpAddr()

	pss.log.Infof("[%v] %v -> /Init( %v, %v, %v )", traceId, sourceInstance,  client, pb.ProxySessionType_name[int32(param.GetSessionType())], stat)

	instance := pss.proxySession.GetInstance()
	if _, err := pss.proxySession.GetProxy().GetSession(client); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, fmt.Sprintf("cannot rename %v to %v: instance already exists", instance, client, err))
	}
	if err := pss.proxySession.SetInstance(client); err != nil {
		return nil, status.Errorf(codes.OutOfRange, fmt.Sprintf("cannot rename %v to %v: %v", instance, client, err))
	}

	pss.proxySession.SetSessionType(common.SessionType(param.SessionType))

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
			Type: pb.ProxySessionType(session.GetSessionType()),
			Status:status,
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

	pss.log.Infof("[%v] %v -> /ws() -> %v", traceId, sourceInstance,  targetGroup)

	instances := pss.proxySession.proxy.groups.GetMembers(targetGroup)
	for _, instanceName := range instances {
		// don't send to myself
		if instanceName == pss.proxySession.GetInstance() {
			continue
		}
		session, err := pss.proxySession.proxy.GetSession(instanceName)
		if err != nil {
			pss.log.Errorf("client %v not active - cannot send websocket message: %v", instanceName, err )
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
