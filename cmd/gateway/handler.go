package main

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/infobloxopen/atlas-contacts-app/pkg/pb"
	"google.golang.org/grpc"
)

// NewAtlasContactsAppHandler returns an HTTP handler that serves the gRPC gateway
func NewAtlasContactsAppHandler(ctx context.Context, grpcAddr string, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	mux := runtime.NewServeMux(append(opts, runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))...)
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterProfilesHandlerFromEndpoint(ctx, mux, grpcAddr, dialOpts)
	if err != nil {
		return nil, err
	}
	err = pb.RegisterGroupsHandlerFromEndpoint(ctx, mux, grpcAddr, dialOpts)
	if err != nil {
		return nil, err
	}
	err = pb.RegisterContactsHandlerFromEndpoint(ctx, mux, grpcAddr, dialOpts)
	if err != nil {
		return nil, err
	}
	return mux, nil
}
