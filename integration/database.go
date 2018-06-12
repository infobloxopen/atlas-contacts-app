package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

// PostgresDBConfig models a test Postgres database
type PostgresDBConfig struct {
	Host            string
	Port            int
	DBName          string
	DBUser          string
	DBPassword      string
	DBVersion       string
	MigrateFunction func(sql.DB) error
}

// NewTestPostgresDB creates a PostgresDBConfig with a sensible set of
// database defaults
func NewTestPostgresDB(dbName string, migrateFunction func(sql.DB) error) (*PostgresDBConfig, error) {
	port, err := GetOpenPortInRange(35000, portRangeMax)
	if err != nil {
		return nil, err
	}
	return &PostgresDBConfig{
		Host:            "localhost",
		Port:            port,
		DBName:          dbName,
		DBUser:          "postgres",
		DBPassword:      "postgres",
		DBVersion:       "latest",
		MigrateFunction: migrateFunction,
	}, nil
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
	if cfg.MigrateFunction != nil {
		if err := cfg.MigrateFunction(*db); err != nil {
			t.Fatalf("unable to migrate %s database: %v", cfg.DBName, err)
		}
	}
}

// RunAsDockerContainer uses the Postgres configuration to run a Postgres Docker
// container on the host machine
func (cfg PostgresDBConfig) RunAsDockerContainer(containerName string) (func() error, error) {
	cleanup, err := RunContainer(
		// define the postgres image version
		fmt.Sprintf("postgres:%s", cfg.DBVersion),
		// define the arguments to docker
		[]string{
			fmt.Sprintf("--name=%s", containerName),
			fmt.Sprintf("--publish=%d:5432", cfg.Port),
			fmt.Sprintf("--env=POSTGRES_DB=%s", cfg.DBName),
			fmt.Sprintf("--env=POSTGRES_PASSWORD=%s", cfg.DBPassword),
			fmt.Sprintf("--env=POSTGRES_USER=%s", cfg.DBUser),
			"--detach",
			"--rm",
		},
		// define the runtime arguments to postgres
		[]string{},
	)
	if err != nil {
		return nil, err
	}

	doneC := make(chan error)
	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// wait until the database container is running and available
	go func() {
		for {
			select {
			case <-time.After(500 * time.Millisecond):
				db, _ := sql.Open("postgres", cfg.GetDSN())
				if err := db.Ping(); err == nil {
					doneC <- nil
					return
				}
			case <-ctx.Done():
				doneC <- fmt.Errorf("failed to start database after %f seconds", timeout.Seconds())
			}
		}
	}()

	if err := <-doneC; err != nil {
		cleanup()
		return nil, err
	}

	return cleanup, nil
}

// GetDSN returns the database connection string for the test Postgres database
func (cfg PostgresDBConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable dbname=%s",
		cfg.Host, cfg.Port, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
}
