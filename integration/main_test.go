// +build integration

package integration

import (
	"log"
	"testing"
	"time"

	migrate "github.com/infobloxopen/atlas-contacts-app/db"
)

const (
	pkgGateway = "../cmd/gateway"
	pkgServer  = "../cmd/server"
)

var (
	dbTest PostgresDBConfig
)

// TestMain launches a gRPC server, REST gateway, and Postgres database
func TestMain(m *testing.M) {
	// initialize the test database
	if db, err := NewTestPostgresDB("contacts", migrate.MigrateDB); err != nil {
		log.Fatalf("unable to create test contacts database: %v", err)
	} else {
		dbTest = *db
	}

	// build the gRPC server binary; delete binary when finished
	log.Printf("building the server binary")
	rmServer, err := BuildSource(pkgServer, "server")
	if err != nil {
		log.Fatalf("failed to build the server: %v", err)
	}
	defer rmServer()

	// build the REST gateway binary; delete binary when finished
	log.Printf("building the gateway binary")
	rmGateway, err := BuildSource(pkgGateway, "gateway")
	if err != nil {
		log.Fatalf("failed to run the gateway: %v", err)
	}
	defer rmGateway()

	// start the gRPC server; stop processes when finished
	log.Printf("running the server binary")
	closeServer, err := RunBinary("server", "-db", dbTest.GetDSN())
	if err != nil {
		log.Fatalf("failed to run the server: %v", err)
	}
	defer closeServer()

	// start the REST gateway; stop processes when finished
	log.Printf("running the gateway binary")
	closeGatway, err := RunBinary("gateway")
	if err != nil {
		log.Fatalf("failed to close the gateway: %v", err)
	}
	defer closeGatway()

	// start the postgres container; stop container when finished
	log.Printf("running the test postgres container")
	closeDatabase, err := dbTest.RunAsDockerContainer("contacts-application-test")
	if err != nil {
		log.Fatalf("failed to launch the test postgres container: %v", err)
	}
	defer closeDatabase()

	// sleep to avoid sending requests to servers that are still starting
	log.Print("waiting for servers to start")
	time.Sleep(time.Second * 2)
	m.Run()
	log.Print("finished running tests. cleaning up now.")
}
