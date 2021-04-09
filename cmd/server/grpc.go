package main

import (
	"github.com/infobloxopen/atlas-app-toolkit/errors"
	"github.com/infobloxopen/atlas-app-toolkit/errors/mappers/validationerrors"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
	grpc_opa "github.com/infobloxopen/atlas-authz-middleware/grpc_opa"
	"github.com/infobloxopen/atlas-contacts-app/pkg/pb"
	"github.com/infobloxopen/atlas-contacts-app/pkg/svc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func NewGRPCServer(logger *logrus.Logger, db *gorm.DB) (*grpc.Server, error) {
	interceptors := []grpc.UnaryServerInterceptor{
		grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logger)),
		requestid.UnaryServerInterceptor(),
		errors.UnaryServerInterceptor(ErrorMappings...),
		// validation interceptor
		validationerrors.UnaryServerInterceptor(),
		gateway.UnaryServerInterceptor(),
	}
	// add authorization interceptor if authz service address is provided
	if AuthzAddr != "" {
		// authorization interceptor
		interceptors = append(interceptors, grpc_opa.UnaryServerInterceptor("contacts"))
	}

	// create new gRPC grpcServer with middleware chain
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)))

	// register all of our services into the grpcServer
	ps, err := svc.NewProfilesServer(db)
	if err != nil {
		return nil, err
	}
	pb.RegisterProfilesServer(grpcServer, ps)

	gs, err := svc.NewGroupsServer(db)
	if err != nil {
		return nil, err
	}
	pb.RegisterGroupsServer(grpcServer, gs)

	cs, err := svc.NewContactsServer(db)
	if err != nil {
		return nil, err
	}
	pb.RegisterContactsServer(grpcServer, cs)

	return grpcServer, nil
}
