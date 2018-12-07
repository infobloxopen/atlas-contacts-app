package main

import (
	"flag"
	"net"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	"fmt"
	"net/http"
	"time"

	"database/sql"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	"github.com/infobloxopen/atlas-app-toolkit/health"
	"github.com/infobloxopen/atlas-app-toolkit/server"
	"github.com/infobloxopen/atlas-contacts-app/cmd"
	"github.com/infobloxopen/atlas-contacts-app/pkg/pb"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"google.golang.org/grpc"
)

var (
	ServerAddress      string
	GatewayAddress     string
	InternalAddress    string
	SwaggerDir         string
	DBConnectionString string
	LogLevel           string
)

func main() {
	doneC := make(chan error)
	logger := NewLogger()

	go func() { doneC <- ServeInternal(logger) }()
	go func() { doneC <- ServeExternal(logger) }()

	if err := <-doneC; err != nil {
		logger.Fatal(err)
	}
}

func init() {
	// default server address; optionally set via command-line flags
	flag.StringVar(&ServerAddress, "address", cmd.ServerAddress, "the gRPC server address")
	flag.StringVar(&GatewayAddress, "gateway", cmd.GatewayAddress, "address of the gateway server")
	flag.StringVar(&InternalAddress, "internal-addr", cmd.InternalAddress, "address of an internal http server, for endpoints that shouldn't be exposed to the public")
	flag.StringVar(&SwaggerDir, "swagger-dir", cmd.SwaggerFile, "directory of the swagger.json file")
	flag.StringVar(&DBConnectionString, "db", cmd.DBConnectionString, "the database address")
	flag.StringVar(&LogLevel, "log", "info", "log level")
	flag.Parse()
	resource.RegisterApplication(cmd.ApplicationID)
}

func NewLogger() *logrus.Logger {
	logger := logrus.StandardLogger()

	// Set the log level on the default logger based on command line flag
	logLevels := map[string]logrus.Level{
		"debug":   logrus.DebugLevel,
		"info":    logrus.InfoLevel,
		"warning": logrus.WarnLevel,
		"error":   logrus.ErrorLevel,
		"fatal":   logrus.FatalLevel,
		"panic":   logrus.PanicLevel,
	}
	if level, ok := logLevels[LogLevel]; !ok {
		logger.Errorf("Invalid value %q provided for log level", LogLevel)
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(level)
	}

	return logger
}

// ServeInternal builds and runs the server that listens on InternalAddress
func ServeInternal(logger *logrus.Logger) error {
	healthChecker := health.NewChecksHandler("/healthz", "/ready")
	healthChecker.AddReadiness("DB ready check", dbReady)
	healthChecker.AddLiveness("ping", health.HTTPGetCheck(fmt.Sprint("http://", InternalAddress, "/ping"), time.Minute))

	s, err := server.NewServer(
		// register our health checks
		server.WithHealthChecks(healthChecker),
		// this endpoint will be used for our health checks
		server.WithHandler("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("pong"))
		})),
	)
	if err != nil {
		return err
	}
	l, err := net.Listen("tcp", InternalAddress)
	if err != nil {
		return err
	}

	logger.Debugf("serving internal http at %q", InternalAddress)

	return s.Serve(nil, l)
}

// ServeExternal builds and runs the server that listens on ServerAddress and GatewayAddress
func ServeExternal(logger *logrus.Logger) error {
	dbSQL, err := sql.Open("postgres", DBConnectionString)
	if err != nil {
		return err
	}
	defer dbSQL.Close()
	db, err := gorm.Open("postgres", dbSQL)
	if err != nil {
		return err
	}
	defer db.Close()

	grpcServer, err := NewGRPCServer(logger, db)
	if err != nil {
		return err
	}

	s, err := server.NewServer(
		// register our grpc server
		server.WithGrpcServer(grpcServer),
		// register the gateway to proxy to the given server address with the service registration endpoints
		server.WithGateway(
			gateway.WithGatewayOptions(
				runtime.WithMetadata(gateway.NewPresenceAnnotator("PUT")),
			),
			gateway.WithDialOptions(
				[]grpc.DialOption{grpc.WithInsecure(), grpc.WithUnaryInterceptor(
					grpc_middleware.ChainUnaryClient(
						[]grpc.UnaryClientInterceptor{gateway.ClientUnaryInterceptor, gateway.PresenceClientInterceptor()}...,
					),
				)}...,
			),
			// TODO: 这里最好还是只取端口号使用，不能直接使用0000
			gateway.WithServerAddress("localhost:9090"),
			gateway.WithEndpointRegistration("/v1/", pb.RegisterProfilesHandlerFromEndpoint, pb.RegisterGroupsHandlerFromEndpoint, pb.RegisterContactsHandlerFromEndpoint),
		),
		// serve swagger at the root
		server.WithHandler("/swagger", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			http.ServeFile(writer, request, SwaggerDir)
		})),
	)
	if err != nil {
		return err
	}

	// open some listeners for our server and gateway
	grpcL, err := net.Listen("tcp", ServerAddress)
	if err != nil {
		return err
	}
	gatewayL, err := net.Listen("tcp", GatewayAddress)
	if err != nil {
		return err
	}

	logger.Debugf("serving gRPC at %q", ServerAddress)
	logger.Debugf("serving http at %q", GatewayAddress)

	return s.Serve(grpcL, gatewayL)
}

func dbReady() error {
	db, err := gorm.Open("postgres", DBConnectionString)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.DB().Ping()
}
