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
	return &server{db: db}, nil
}

type server struct {
	db *gorm.DB
}

func (s *server) Create(ctx context.Context, req *pb.CreateContactRequest) (*pb.CreateContactResponse, error) {
	resp := &pb.CreateContactResponse{}
	result, err := pb.DefaultCreateContact(ctx, req.GetPayload(), s.db)
	resp.Result = result
	return resp, err
}

func (s *server) Get(ctx context.Context, req *pb.GetContactRequest) (*pb.GetContactResponse, error) {
	resp := &pb.GetContactResponse{}
	result, err := pb.DefaultReadContact(ctx, &pb.Contact{Id: req.GetId()}, s.db)
	resp.Result = result
	return resp, err
}

func (s *server) Update(ctx context.Context, req *pb.UpdateContactRequest) (*pb.UpdateContactResponse, error) {
	resp := &pb.UpdateContactResponse{}
	result, err := pb.DefaultUpdateContact(ctx, req.GetPayload(), s.db)
	resp.Result = result
	return resp, err
}

func (s *server) Delete(ctx context.Context, req *pb.DeleteContactRequest) (*google_protobuf.Empty, error) {
	return &google_protobuf.Empty{}, pb.DefaultDeleteContact(ctx, &pb.Contact{Id: req.GetId()}, s.db)
}

func (s *server) List(ctx context.Context, req *google_protobuf.Empty) (*pb.ListContactsResponse, error) {
	resp := &pb.ListContactsResponse{}
	results, err := pb.DefaultListContact(ctx, s.db)
	resp.Results = results
	return resp, err
}

func (s *server) SendSMS(context.Context, *pb.SMSRequest) (*google_protobuf.Empty, error) {
	// suppose this method wanted to send the message then update a field in the contact
	// in that case it could begin a transaction, and do a select for update
	//   s.db.Begin(), etc.
	//   try to send the message and log the result in another table
	//   and commit/rollback as needed; it's up to the app

	return &google_protobuf.Empty{}, nil
}
