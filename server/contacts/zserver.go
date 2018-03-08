package contacts

import (
	"context"

	google_protobuf "github.com/golang/protobuf/ptypes/empty"
        "github.com/jinzhu/gorm"
        _ "github.com/jinzhu/gorm/dialects/mysql"

	pb "github.com/infobloxopen/atlas-contacts-app/pb/contacts"
	orm "github.com/infobloxopen/atlas-contacts-app/orm/contacts"
)


// CreateContact will actually live in the generated ORM code
// But for now it is here to give an example of how the basic
// server will use it, and then how you can extend that
func CreateContact(db *gorm.DB, c *pb.Contact) (*pb.Contact, error) {
	dbc := orm.ConvertToContact(*c)
        if err:= db.Create(&dbc).Error; err != nil {
		return nil, err
        }

	cc := orm.ConvertFromContact(dbc)
	return &cc, nil
}

func NewBasicServer(dsn string) (pb.ContactsServer, error) {
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

// Create would use the generated method to satisfy the basic operation
func (s *server) Create(ctx context.Context, c *pb.Contact) (*pb.Contact, error) {
	return CreateContact(s.db, c)
}

// Search would also use a generated method
func (s *server) Search(context.Context, *pb.SearchRequest) (*pb.ContactPage, error) {
	return &pb.ContactPage{}, nil
}

// Get would also use a generated method
func (s *server) Get(context.Context, *pb.GetRequest) (*pb.Contact, error) {
	return &pb.Contact{}, nil
}

// Get would also use a generated method
func (s *server) Update(context.Context, *pb.Contact) (*pb.Contact, error) {
	return &pb.Contact{}, nil
}

// Get would also use a generated method
func (s *server) Delete(context.Context, *pb.GetRequest) (*google_protobuf.Empty, error) {
	return &google_protobuf.Empty{}, nil
}

// SendSMS would not be generated as it is a non-CRUD method

