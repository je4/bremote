package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/je4/bremote/api"
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

	pss.log.Infof("[%v] %v -> /Init( %v, %v, %v )", traceId, sourceInstance,  client, pb.ProxySessionType_name[int32(param.GetSessionType())], stat)

	instance := pss.proxySession.GetInstance()
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
			cw := api.NewControllerWrapper(name, session.GetSessionPtr())

			traceId = uniqid.New(uniqid.Params{"traceid_", false})
			pss.log.Infof("[%v] announcing %v to %v", traceId, client, name)
			if err := cw.NewClient(traceId, name, client, stat); err != nil {
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
