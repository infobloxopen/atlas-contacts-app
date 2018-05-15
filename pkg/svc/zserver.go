package svc

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	pb "github.com/infobloxopen/atlas-contacts-app/pkg/pb"
)

// NewBasicServer constructs a new BasicServer and connects to a postgres db
func NewBasicServer(db *gorm.DB) (pb.ContactsServer, error) {
	return &server{&pb.ContactsDefaultServer{db}}, nil
}

type server struct {
	*pb.ContactsDefaultServer
}
