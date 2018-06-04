package integration

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"testing"
)

// PostgresDBConfig models a test Postgres database
type PostgresDBConfig struct {
	Host            string
	Port            int
	DBUser          string
	DBName          string
	DBPassword      string
	migrateFunction func(sql.DB) error
}

// Reset drops all the tables in a test database and regenerates them by
// running migrations. If a migrationFunction has not been specified, then the
// tables are dropped but not regenerated
func (cfg PostgresDBConfig) Reset(t *testing.T) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		t.Fatalf("unable to connect to %s database: %v", cfg.DBName, err)
	}
	defer db.Close()
	resetQuery := "DROP SCHEMA public CASCADE;" +
		"CREATE SCHEMA public;" +
		"GRANT ALL ON SCHEMA public TO postgres;" +
		"GRANT ALL ON SCHEMA public TO public;"
	// drop all the tables in the test database
	if _, err := db.Exec(resetQuery); err != nil {
		t.Fatalf("unable to drop tables in %s database: %v", cfg.DBName, err)
	}
	// run migrations if a migration function has exists
	if cfg.migrateFunction != nil {
		if err := cfg.migrateFunction(*db); err != nil {
			t.Fatalf("unable to migrate %s database: %v", cfg.DBName, err)
		}
	}
}

// RunAsContainer uses the Postgres configuration to run a Postgres Docker
// container locally
func (cfg PostgresDBConfig) RunAsContainer(containerName string) func() {
	log.Printf("starting the %s postgres container", containerName)
	args := []string{
		"run", "--rm", "-d",
		"--name", containerName,
		"-p", fmt.Sprintf("%d:5432", cfg.Port),
		"-e", fmt.Sprintf("POSTGRES_DB=%s", cfg.DBName),
		"postgres:latest",
	}
	cmdRunContainer := exec.Command("docker", args...)
	if err := cmdRunContainer.Run(); err != nil {
		log.Fatalf("unable to start test database: %v", err)
	}
	return func() {
		cmdStopContainer := exec.Command("docker", "kill", containerName)
		log.Printf("stopping the %s postgres container", containerName)
		if err := cmdStopContainer.Run(); err != nil {
			log.Fatalf("unable to stop test database: %v", err)
		}
	}
}

// GetDSN returns the database connection string for the test Postgres database
func (cfg PostgresDBConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable dbname=%s",
		cfg.Host, cfg.Port, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
}
