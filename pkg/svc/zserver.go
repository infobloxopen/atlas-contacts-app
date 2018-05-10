package svc

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	pb "github.com/infobloxopen/atlas-contacts-app/pkg/pb"
)

// NewBasicServer constructs a new BasicServer and connects to a postgres db
func NewBasicServer(dsn string) (pb.ContactsServer, error) {
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// This string creates table for Contact objects.
	// In case you didn't have one already.
	// This is done to make current example up and running
	// without having contacts DB configured properly.
	//if err := db.AutoMigrate(&pb.Contact{}).Error; err != nil {
	//	return nil, err
	//}

	return &server{&pb.ContactsDefaultServer{db}}, nil
}

type server struct {
	*pb.ContactsDefaultServer
}
