package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/goph/emperror"
	"github.com/hashicorp/yamux"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	pb "info_age.net/bremote/api"
	"net"
)

type Client struct {
	log        *logging.Logger
	instance string
	conn       *tls.Conn
	session    *yamux.Session
	grpcServer *grpc.Server
}

func NewClient(instance string, log *logging.Logger) *Client {
	client := &Client{log: log, instance:instance}
	return client
}

func (client *Client) Connect() (err error) {

	// First, create the set of root certificates. For this example we only
	// have one. It's also possible to omit this in order to use the
	// default root set of the current operating system.
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("failed to parse root certificate")
	}

	client.log.Infof("trying to connect %v", addr)
	client.conn, err = tls.Dial("tcp", addr, &tls.Config{
		RootCAs:            roots,
		InsecureSkipVerify: true,
		ServerName:         "localhost",
	})
	if err != nil {
		return emperror.Wrapf(err, "cannot connect to %v", addr)
	}
	client.log.Info("connection established")

	client.session, err = yamux.Server(client.conn, yamux.DefaultConfig())
	if err != nil {
		return emperror.Wrap(err, "cannot setup yamux client")
	}
	client.log.Info("yamux session established")

	return
}

func (client *Client) InitProxy() error {

	// gRPC dial over incoming net.Conn
	conn, err := grpc.Dial(":7777", grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return client.session.Open()
		}),
	)
	if err != nil {
		return errors.New("cannot dial grpc connection to :7777")
	}
	proxy := pb.NewProxyServiceClient(conn)
	str := &pb.InitParam{Instance:&pb.String{Value:client.instance},
		SessionType:pb.ProxySessionType_Client}
	_, err = proxy.Init(context.Background(), str)
	if err != nil {
		return emperror.Wrap(err, "cannot initialize client")
	}
	return nil
}

func (client *Client) ServeGRPC() error {
	// create a server instance
	server := NewClientServiceServer(client.log)

	// create a gRPC server object
	client.grpcServer = grpc.NewServer()

	// attach the Ping service to the server
	pb.RegisterClientServiceServer(client.grpcServer, &server)

	// start the gRPC erver
	client.log.Info("launching gRPC server over TLS connection...")
	if err := client.grpcServer.Serve(client.session); err != nil {
		return emperror.Wrapf(err, "failed to serve")
	}
	return nil
}

func (client *Client) Close() error {
	client.grpcServer.GracefulStop()
	client.session.Close()
	client.conn.Close()
	return nil
}
