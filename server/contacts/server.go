package contacts

import (
	"context"

	google_protobuf "github.com/golang/protobuf/ptypes/empty"
        "github.com/jinzhu/gorm"
        _ "github.com/jinzhu/gorm/dialects/mysql"

	pb "github.com/infobloxopen/atlas-contacts-app/pb/contacts"
	orm "github.com/infobloxopen/atlas-contacts-app/orm/contacts"
)

func NewServer(dsn string) (pb.ContactsServer, error) {
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &server{dsn: dsn, db: db}, nil
}

type server struct {
	dsn string
	db *gorm.DB
}

func (s *server) Search(context.Context, *pb.SearchRequest) (*pb.ContactPage, error) {
	return &pb.ContactPage{}, nil
}

func (s *server) Create(ctx context.Context, c *pb.Contact) (*pb.Contact, error) {
	dbc := orm.ConvertToContact(*c)
        if err:= s.db.Create(&dbc).Error; err != nil {
		return nil, err
        }

	cc := orm.ConvertFromContact(dbc)
	return &cc, nil
}

func (s *server) Get(context.Context, *pb.SearchCriteria) (*pb.Contact, error) {
	return &pb.Contact{}, nil
}

func (s *server) Update(context.Context, *pb.Contact) (*pb.Contact, error) {
	return &pb.Contact{}, nil
}

func (s *server) Delete(context.Context, *pb.SearchCriteria) (*google_protobuf.Empty, error) {
	return &google_protobuf.Empty{}, nil
}
