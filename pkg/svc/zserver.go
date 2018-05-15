package svc

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	pb "github.com/infobloxopen/atlas-contacts-app/pkg/pb"
)

// NewBasicServer returns an instance of the default server interface
func NewBasicServer(database *gorm.DB) (pb.ContactsServer, error) {
	return &pb.ContactsDefaultServer{DB: database}, nil
}

type server struct {
	*pb.ContactsDefaultServer
}
