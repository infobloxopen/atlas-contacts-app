package main

import (
	"flag"
	"net"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"
	toolkit_auth "github.com/infobloxopen/atlas-app-toolkit/mw/auth"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/infobloxopen/atlas-app-toolkit/health"
	"github.com/infobloxopen/atlas-app-toolkit/pkg/appserver"
	atflag "github.com/infobloxopen/atlas-app-toolkit/pkg/flag"
	pb "github.com/infobloxopen/atlas-contacts-app/pb/contacts"
	svc "github.com/infobloxopen/atlas-contacts-app/svc/contacts"
)

var (
	InitTimeout int64
	Dsn         string
	AuthzAddr   string
)

func main() {
	logger := logrus.New()
	// create new tcp listenerf
	ln, err := net.Listen("tcp", Address)
	if err != nil {
		logger.Fatalln(err)
	}

	interceptors := []grpc.UnaryServerInterceptor{
		// validation interceptor
		grpc_validator.UnaryServerInterceptor(),
		// validation interceptor
		grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logger)),
	}
	// add authorization interceptor if authz service address is provided
	if AuthzAddr != "" {
		interceptors = append(interceptors,
			// authorization interceptor
			toolkit_auth.DefaultAuthInterceptor(AuthzAddr),
		)
	}
	middleware := grpc_middleware.ChainUnaryServer(interceptors...)

	// create new gRPC server with middleware chain
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			middleware,
		),
	)
	// define an initializer function for server initialization
	initializer := func() error {
		// register service implementation with the grpc server
		s, err := svc.NewBasicServer(Dsn)
		if err != nil {
			return err
		}
		pb.RegisterContactsServer(server, s)
		return nil
	}

	healthFlags := atflag.NewHealthProbesFlags()
	grpcFlags := atflag.NewGRPCFlags()
	flag.Parse()

	var opts []appserver.Option

	// TODO: Provide mechanism for determination of what flags user set
	// and what he didn't. This is needed to determine whether to include particular option or not.
	// Simplest method is to remove default values for corresponding flags.
	healthChecker := health.NewChecksHandler(healthFlags.HealthPath(), healthFlags.ReadyPath())

	opts = append(opts, appserver.WithHealthOptions(healthFlags.Addr(), healthChecker.Handler()))
	opts = append(opts, appserver.WithGRPC(grpcFlags.Addr(), server))

	initOption := appserver.WithInitializer(initializer, time.Duration(InitTimeout)*time.Second)
	opts = append(opts, initOption)

	appServer, err := appserver.NewAppServer(opts...)
	if err != nil {
		logger.Fatalln(err)
	}

	if err = appServer.Serve(); err != nil {
		logger.Fatalln(err)
	}
}

func init() {
	// default server address; optionally set via command-line flags
	flag.StringVar(&Dsn, "dsn", "", "")
	flag.StringVar(&AuthzAddr, "authz", "", "address of the authorization service")
	flag.Int64Var(&InitTimeout, "init-timeout", 10,
		"Timeout in seconds needed for initialization procedure to pass")
}

func dbReady() error {
	db, err := gorm.Open("postgres", Dsn)
	if err != nil {
		return err
	}
	db.Close()
	return nil
}
