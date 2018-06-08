package main

import (
	"context"
	"database/sql"
	"flag"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"
	toolkit_auth "github.com/infobloxopen/atlas-app-toolkit/auth"
	migrate "github.com/infobloxopen/atlas-contacts-app/db"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/infobloxopen/atlas-app-toolkit/health"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
	"github.com/infobloxopen/atlas-contacts-app/cmd"
	"github.com/infobloxopen/atlas-contacts-app/pkg/pb"
	"github.com/infobloxopen/atlas-contacts-app/pkg/svc"
)

var (
	ServerAddress      string
	HealthAddress      string
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

	// create new tcp listenerf
	ln, err := net.Listen("tcp", ServerAddress)
	if err != nil {
		logger.Fatalln(err)
	}

	interceptors := []grpc.UnaryServerInterceptor{
		// validation interceptor
		grpc_validator.UnaryServerInterceptor(),
		// logger interceptor
		grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logger)),
		// Request-Id interceptor
		requestid.UnaryServerInterceptor(),
	}
	// add authorization interceptor if authz service address is provided
	if AuthzAddr != "" {
		interceptors = append(interceptors,
			// authorization interceptor
			toolkit_auth.UnaryServerInterceptor(AuthzAddr, applicationID),
		)
	}
	middleware := grpc_middleware.ChainUnaryServer(interceptors...)
	// create new gRPC server with middleware chain
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			middleware,
		),
	)

	healthChecker := health.NewChecksHandler("/healthz", "/ready")
	healthChecker.AddReadiness("DB ready check", dbReady)
	go http.ListenAndServe(HealthAddress, healthChecker.Handler())

	// waiting for database is available.
	logger.Info("Connecting to database...")
	dbCheckContext, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	for {
		select {
		case <-time.After(time.Second * 3):
			if err = dbReady(); err != nil {
				continue
			}
			cancel()
		case <-dbCheckContext.Done():
			cancel()
		}
		break
	}
	if err != nil {
		logger.Fatalln(err)
	}
	logger.Info("Connected to database successfully.")
	dbSQL, err := sql.Open("postgres", DBConnectionString)
	if err != nil {
		logger.Fatal(err)
	}

	defer dbSQL.Close()
	if err := migrate.MigrateDB(*dbSQL); err != nil {
		logger.Fatal(err)
	}
	db, err := gorm.Open("postgres", dbSQL)
	if err != nil {
		logger.Fatalln(err)
	}
	ps, err := svc.NewProfilesServer(db)
	if err != nil {
		logger.Fatalln(err)
	}
	pb.RegisterProfilesServer(server, ps)

	gs, err := svc.NewGroupsServer(db)
	if err != nil {
		logger.Fatalln(err)
	}
	pb.RegisterGroupsServer(server, gs)

	cs, err := svc.NewContactsServer(db)
	if err != nil {
		logger.Fatalln(err)
	}
	pb.RegisterContactsServer(server, cs)

	if err := server.Serve(ln); err != nil {
		logger.Fatalln(err)
	}
}

func init() {
	// default server address; optionally set via command-line flags
	flag.StringVar(&ServerAddress, "address", cmd.ServerAddress, "the gRPC server address")
	flag.StringVar(&DBConnectionString, "db", "", "the database address")
	flag.StringVar(&HealthAddress, "health", "0.0.0.0:8089", "Address for health checking")
	flag.StringVar(&AuthzAddr, "authz", "", "address of the authorization service")
	flag.StringVar(&LogLevel, "log", "info", "log level")
	flag.Parse()
}

func dbReady() error {
	db, err := gorm.Open("postgres", DBConnectionString)
	if err != nil {
		return err
	}
	db.Close()
	return nil
}
