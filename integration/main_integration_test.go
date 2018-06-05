// +build integration

package integration

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	migrate "github.com/infobloxopen/atlas-contacts-app/db"
)

const (
	pkgGateway = "../cmd/gateway"
	pkgServer  = "../cmd/server"
)

var (
	dbTest = PostgresDBConfig{
		Host:            "localhost",
		Port:            5425,
		DBUser:          "postgres",
		DBPassword:      "postgres",
		DBName:          "contacts",
		migrateFunction: migrate.MigrateDB,
	}
)

// TestMain launches a gRPC server, REST gateway, and Postgres database
func TestMain(m *testing.M) {
	// start the test postgres container; stop the container when finished
	defer dbTest.RunAsContainer("contacts-application-test")()

	// build the gateway and server binaries; delete binaries when finished
	defer BuildSource(pkgServer, "server")()
	defer BuildSource(pkgGateway, "gateway")()

	// run the gateway and server binaries; stop processes when finished
	defer RunBinary("server", "-db", dbTest.GetDSN())()
	defer RunBinary("gateway")()

	// sleep to avoid sending requests to servers that are still starting
	log.Print("waiting for servers to start")
	time.Sleep(time.Second * 1)

	m.Run()
}

// RunBinary runs a target binary with a set of arguments provided by the user
func RunBinary(binPath string, args ...string) func() {
	log.Printf("running the %s binary", binPath)
	abs, err := filepath.Abs(binPath)
	if err != nil {
		log.Fatalf("unable to get absolute path of the %s binary: %v", binPath, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := exec.CommandContext(ctx, abs, args...).Start(); err != nil {
		log.Fatalf("unable to start the %s binary: %v", binPath, err)
	}
	return func() {
		log.Printf("stopping %s process", binPath)
		cancel()
	}
}

// BuildSource builds a target Go package and gives the resulting binary
// some user-defined name
func BuildSource(packagePath, output string) func() {
	log.Printf("creating %s binary", output)
	cmdBuild := exec.Command("go", "build", "-o", output, packagePath)
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		log.Fatalf("unable to build the %s package: %v (%s)",
			packagePath, err, string(out),
		)
	}
	return func() {
		if err := os.Remove(output); err != nil {
			log.Fatalf("unable to delete the %s binary: %v", output, err)
		}
		log.Printf("deleting %s binary", output)
	}
}
