// +build integration

package integration

import (
	"log"
	"testing"
	"time"

	migrate "github.com/infobloxopen/atlas-contacts-app/db"
)

const (
	pkgServer = "../cmd/server"
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

	// start the postgres container; stop container when finished
	log.Printf("running the test postgres container")
	closeDatabase, err := dbTest.RunAsDockerContainer("contacts-application-test")
	if err != nil {
		log.Fatalf("failed to launch the test postgres container: %v", err)
	}
	defer closeDatabase()

	// build the gRPC server binary; delete binary when finished
	log.Printf("building the server binary")
	rmServer, err := BuildSource(pkgServer, "server")
	if err != nil {
		log.Fatalf("failed to build the server: %v", err)
	}
	defer rmServer()

	// start the gRPC server; stop processes when finished
	log.Printf("running the server binary")
	closeServer, err := RunBinary("server", "-db", dbTest.GetDSN())
	if err != nil {
		log.Fatalf("failed to run the server: %v", err)
	}
	defer closeServer()

	// sleep to avoid sending requests to servers that are still starting
	log.Print("waiting for servers to start")
	time.Sleep(time.Second * 2)
	m.Run()
	log.Print("finished running tests. cleaning up now.")
}
