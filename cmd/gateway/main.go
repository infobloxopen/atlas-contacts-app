package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/infobloxopen/atlas-app-toolkit/gw"
	"github.com/infobloxopen/atlas-app-toolkit/health"
	"github.com/infobloxopen/atlas-contacts-app/cmd"
)

const readinessTimeout = time.Second * 10

var (
	ServerAddress  string
	GatewayAddress string
	SwaggerDir     string
	HealthAddress  string
	ServerHealth   string
)

func main() {
	// create HTTP handler for gateway
	errHandler := runtime.WithProtoErrorHandler(gw.ProtoMessageErrorHandler)
	opHandler := runtime.WithMetadata(gw.MetadataAnnotator)
	serverHandler, err := NewAtlasContactsAppHandler(context.Background(), ServerAddress, errHandler, opHandler)
	// strip all but trailing "/" on incoming requests
	serverHandler = http.StripPrefix(
		cmd.GatewayURL[:len(cmd.GatewayURL)-1],
		serverHandler,
	)
	if err != nil {
		log.Fatalln(err)
	}
	// map HTTP endpoints to handlers
	mux := http.NewServeMux()
	mux.Handle("/atlas-contacts-app/v1/", serverHandler)
	mux.HandleFunc("/swagger/", SwaggerHandler)

	healthChecker := health.NewChecksHandler("/healthz", "/ready")
	healthChecker.AddReadiness("app ready check", isAppReady)
	go http.ListenAndServe(HealthAddress, healthChecker.Handler())

	// serve handlers on the gateway address
	http.ListenAndServe(GatewayAddress, mux)
}

func init() {
	// default gateway values; optionally configured via command-line flags
	flag.StringVar(&ServerAddress, "server", cmd.ServerAddress, "address of the gRPC server")
	flag.StringVar(&HealthAddress, "health", cmd.HEALTH_ADDRESS, "address of readiness check endpoint")
	flag.StringVar(&ServerHealth, "shealth", cmd.SERVER_HEALTH, "address of readiness check endpoint")
	flag.StringVar(&GatewayAddress, "gateway", cmd.GatewayAddress, "address of the gateway server")
	flag.StringVar(&SwaggerDir, "swagger-dir", cmd.SWAGGER_DIR, "directory of the swagger.json file")
	flag.Parse()
}

func isAppReady() error {
	client := &http.Client{
		Timeout: readinessTimeout,
	}

	resp, err := client.Get("http://" + path.Join(ServerHealth, "ready"))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Readiness check failed: %s", resp.Status)
	}
	return nil
}
