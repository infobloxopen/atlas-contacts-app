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
	mux := runtime.NewServeMux(opts...)
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterContactsHandlerFromEndpoint(ctx, mux, grpcAddr, dialOpts)
	if err != nil {
		return nil, err
	}
	return mux, nil
}
