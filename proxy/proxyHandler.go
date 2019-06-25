package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/op/go-logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "info_age.net/bremote/api"
)

type ProxyServiceServer struct {
	log      *logging.Logger
	proxySession *ProxySession
}

func NewProxyServiceServer(proxySession *ProxySession, log *logging.Logger) *ProxyServiceServer {
	pss := &ProxyServiceServer{proxySession:proxySession, log: log}
	return pss
}

func (pss ProxyServiceServer) Ping(ctx context.Context, param *pb.String) (*pb.String, error) {
	pss.log.Infof("Ping( %v )", param.GetValue())

	ret := new(pb.String)
	ret.Value = "pong: " + param.GetValue()

	return ret, nil
}

func (pss ProxyServiceServer) Init(ctx context.Context, param *pb.InitParam) (*empty.Empty, error) {
	instance := pss.proxySession.GetInstance()
	if err := pss.proxySession.SetInstance(param.GetInstance().GetValue()); err != nil {
		return nil, status.Errorf(codes.OutOfRange, fmt.Sprintf("cannot rename %v to %v: %v", instance, param.GetInstance().GetValue(), err))
	}

	switch param.SessionType {
	case pb.ProxySessionType_Undefined:
		pss.proxySession.SetSessionType(SessionType_Undefined)
	case pb.ProxySessionType_Client:
		pss.proxySession.SetSessionType(SessionType_Client)
	case pb.ProxySessionType_Controller:
		pss.proxySession.SetSessionType(SessionType_Controller)
	default:
		return nil, status.Errorf(codes.OutOfRange, fmt.Sprintf("invalid sessiontype %v", param.SessionType))
	}

	return new(empty.Empty), nil
}

func (pss ProxyServiceServer) GetClients(ctx context.Context, req *empty.Empty) (*pb.ProxyClientList, error) {
	clients := new(pb.ProxyClientList)
	clients.Clients = []*pb.ProxyClient{}
	for name, session := range pss.proxySession.GetProxy().GetSessions() {
		// ignore yourself
		if name == pss.proxySession.GetInstance() {
			continue
		}

		clients.Clients = append(clients.Clients, &pb.ProxyClient{Instance:session.GetInstance(), Type:pb.ProxySessionType(session.GetSessionType())})
	}
	return clients, nil
}