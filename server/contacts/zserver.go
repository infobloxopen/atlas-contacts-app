package contacts

import (
	"golang.org/x/net/context"

	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	pb "github.com/infobloxopen/atlas-contacts-app/pb/contacts"
)

// NewBasicServer constructs a new BasicServer and connects to a mysql db
func NewBasicServer(dsn string) (pb.ContactsServer, error) {
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &server{dsn: dsn, db: db}, nil
}

type server struct {
	dsn string
	db  *gorm.DB
}

// Create would use the generated method to satisfy the basic operation
func (s *server) Create(ctx context.Context, c *pb.Contact) (*pb.Contact, error) {
	return pb.DefaultCreateContact(ctx, c, s.db)
}

// Search would also use a generated method
func (s *server) Search(context.Context, *pb.SearchRequest) (*pb.ContactPage, error) {
	return &pb.ContactPage{}, nil
}

// Get would also use a generated method
func (s *server) Get(ctx context.Context, r *pb.GetRequest) (*pb.Contact, error) {
	return pb.DefaultReadContact(ctx, &pb.Contact{Id: r.GetId()}, s.db)
}

// Get would also use a generated method
func (s *server) Update(ctx context.Context, c *pb.Contact) (*pb.Contact, error) {
	return pb.DefaultUpdateContact(ctx, c, s.db)
}

// Get would also use a generated method
func (s *server) Delete(context.Context, *pb.GetRequest) (*google_protobuf.Empty, error) {
	return &google_protobuf.Empty{}, nil
}

// SendSMS would not be generated as it is a non-CRUD method
