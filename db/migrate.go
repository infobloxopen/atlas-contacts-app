package db

import (
	"database/sql"

	"github.com/infobloxopen/atlas-contacts-app/pkg/pb"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// &pb.ProfileORM{}, &pb.GroupORM{}, &pb.ContactORM{}, &pb.AddressORM{}, &pb.EmailORM{}

// MigrateDB builds the contacts application database tables
func MigrateDB(dbSQL sql.DB) error {
	db, err := gorm.Open("postgres", &dbSQL)
	if err != nil {
		return err
	}
	defer db.Close()
	// NOTE: Using db.AutoMigrate is a temporary measure to structure the contacts
	// database schema. The atlas-app-toolkit team will come up with a better
	// solution that uses database migration files.
	return db.AutoMigrate(
		&pb.ProfileORM{}, &pb.GroupORM{}, &pb.ContactORM{}, &pb.AddressORM{}, &pb.EmailORM{},
	).Error
}
