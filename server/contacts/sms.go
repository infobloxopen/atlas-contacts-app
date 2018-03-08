package contacts

import (
	"context"

	google_protobuf "github.com/golang/protobuf/ptypes/empty"

	pb "github.com/infobloxopen/atlas-contacts-app/pb/contacts"
)


// SendSMS is a hand-crafted method that adds onto the BasicServer
// if this is not implemented then the server won't compile
func (s *server) SendSMS(context.Context, *pb.SMSRequest) (*google_protobuf.Empty, error) {
	// suppose this method wanted to send the message then update a field in the contact
	// in that case it could begin a transaction, and do a select for update
	//   s.db.Begin(), etc.
	//   try to send the message and log the result in another table
	//   and commit/rollback as needed; it's up to the app

	return &google_protobuf.Empty{}, nil
}
