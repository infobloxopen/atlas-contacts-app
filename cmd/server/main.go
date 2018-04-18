package main

import (
	"flag"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/infobloxopen/atlas-contacts-app/cmd/config"
	pb "github.com/infobloxopen/atlas-contacts-app/pb/contacts"
	svc "github.com/infobloxopen/atlas-contacts-app/svc/contacts"
)

var (
	Address string
	Dsn     string
)

func main() {
	logger := logrus.New()
	// create new tcp listenerf
	ln, err := net.Listen("tcp", Address)
	if err != nil {
		logger.Fatalln(err)
	}
	// create new gRPC server with middleware chain
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				// validation middleware
				grpc_validator.UnaryServerInterceptor(),
				// logging middleware
				grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logger)),
			),
		),
	)
	// register service implementation with the grpc server
	s, err := svc.NewBasicServer(Dsn)
	if err != nil {
		logger.Fatalln(err)
	}
	pb.RegisterContactsServer(server, s)
	if err := server.Serve(ln); err != nil {
		logger.Fatalln(err)
	}
}

func init() {
	// default server address; optionally set via command-line flags
	flag.StringVar(&Address, "address", config.SERVER_ADDRESS, "the gRPC server address")
	flag.StringVar(&Dsn, "dsn", "", "")
	flag.Parse()
}
