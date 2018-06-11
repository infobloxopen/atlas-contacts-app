package main

import (
	"flag"
	"net"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/health"
	"github.com/infobloxopen/atlas-app-toolkit/server"
	"github.com/infobloxopen/atlas-contacts-app/cmd"
	"github.com/infobloxopen/atlas-contacts-app/pkg/pb"
)

var (
	ServerAddress      string
	GatewayAddress     string
	SwaggerDir         string
	DBConnectionString string
	AuthzAddr          string
	LogLevel           string
)

const (
	// applicationID associates a microservice with an application. the atlas
	// contacts application consists of only one service, so we identify both the
	// service and the application as "contacts"
	applicationID = "contacts"
)

func main() {
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

	// create new postgres database
	db, err := gorm.Open("postgres", DBConnectionString)
	if err != nil {
		logger.Fatalln(err)
	}

	grpcServer, err := NewGRPCServer(logger, db)
	if err != nil {
		logger.Fatalln(err)
	}

	healthChecker := health.NewChecksHandler("/healthz", "/ready")
	healthChecker.AddReadiness("DB ready check", dbReady)
	healthChecker.AddLiveness("ping", health.HTTPGetCheck(fmt.Sprint("http://", GatewayAddress, "/ping"), time.Minute))

	s, err := server.NewServer(
		// upon startup, migrate the database
		server.WithInitializer(func(context.Context) error {
			// NOTE: Using db.AutoMigrate is a temporary measure to structure the contacts
			// database schema. The atlas-app-toolkit team will come up with a better
			// solution that uses database migration files.
			logger.Print("migrating database...")
			defer logger.Print("finished migrating database")
			return db.AutoMigrate(&pb.ProfileORM{}, &pb.GroupORM{}, &pb.ContactORM{}, &pb.AddressORM{}, &pb.EmailORM{}).Error
		}),
		// register our grpc server
		server.WithGrpcServer(grpcServer),
		// register the gateway to proxy to the given server address with the service registration endpoints
		server.WithGateway(
			gateway.WithServerAddress(ServerAddress),
			gateway.WithEndpointRegistration("/v1/", pb.RegisterProfilesHandlerFromEndpoint, pb.RegisterGroupsHandlerFromEndpoint, pb.RegisterContactsHandlerFromEndpoint),
		),
		// register our health checks
		server.WithHealthChecks(healthChecker),
		// serve swagger at the root
		server.WithHandler("/swagger", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			http.ServeFile(writer, request, SwaggerDir)
		})),
		// this endpoint will be used for our health checks
		server.WithHandler("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("pong"))
		})),
	)
	if err != nil {
		logger.Fatalln(err)
	}

	// open some listeners for our server and gateway
	grpcL, err := net.Listen("tcp", ServerAddress)
	if err != nil {
		logger.Fatalln(err)
	}
	httpL, err := net.Listen("tcp", GatewayAddress)
	if err != nil {
		logger.Fatalln(err)
	}

	logger.Printf("serving gRPC at %q", ServerAddress)
	logger.Printf("serving http at %q", GatewayAddress)

	if err := s.Serve(grpcL, httpL); err != nil {
		logger.Fatalln(err)
	}
}

func init() {
	// default server address; optionally set via command-line flags
	flag.StringVar(&ServerAddress, "address", cmd.ServerAddress, "the gRPC server address")
	flag.StringVar(&GatewayAddress, "gateway", cmd.GatewayAddress, "address of the gateway server")
	flag.StringVar(&SwaggerDir, "swagger-dir", cmd.SwaggerFile, "directory of the swagger.json file")
	flag.StringVar(&DBConnectionString, "db", cmd.DBConnectionString, "the database address")
	flag.StringVar(&AuthzAddr, "authz", "", "address of the authorization service")
	flag.StringVar(&LogLevel, "log", "info", "log level")
	flag.Parse()
}

func dbReady() error {
	db, err := gorm.Open("postgres", DBConnectionString)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.DB().Ping()
}
