package main

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/atlas-contacts-app/pb/contacts"
)

// NewAtlasContactsAppHandler returns an HTTP handler that serves the gRPC gateway
func NewAtlasContactsAppHandler(ctx context.Context, addr string, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	mux := runtime.NewServeMux(opts...)
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterContactsHandlerFromEndpoint(ctx, mux, addr, dialOpts)
	if err != nil {
		return nil, err
	}
	return mux, nil
}
