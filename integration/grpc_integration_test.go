// +build integration

package integration

import (
	"testing"

	"github.com/infobloxopen/atlas-contacts-app/cmd"
	"github.com/infobloxopen/atlas-contacts-app/pkg/pb"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"google.golang.org/grpc"
)

func newContactsClient(t *testing.T) (pb.ContactsClient, func()) {
	conn, err := grpc.Dial(cmd.ServerAddress, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("unable to connect to server: %v", err)
	}
	return pb.NewContactsClient(conn), func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("unable to close client: %v", err)
		}
	}
}

func TestCreateContact(t *testing.T) {
	dbTest.Reset(t)
	client, close := newContactsClient(t)
	defer close()
	payload := &pb.Contact{
		FirstName:    "Gandalf",
		MiddleName:   "The",
		LastName:     "Grey",
		PrimaryEmail: "local-wizard@shire.com",
		Emails: []*pb.Email{
			{
				Address: "gandalf@middle-earth.net",
			},
		},
	}
	resCreate, err := client.Create(DefaultContext(t), &pb.CreateContactRequest{
		Payload: payload,
	})
	if err != nil {
		t.Fatalf("unable to create new contact: %s", err)
	}
	resRead, err := client.Read(DefaultContext(t), &pb.ReadContactRequest{
		Id: resCreate.GetResult().GetId(),
	})
	if err != nil {
		t.Fatalf("unable to read contact: %s", err)
	}
	if resRead.GetResult().GetFirstName() != payload.GetFirstName() {
		t.Fatalf("unexpected contact first name: have %s; expected %s",
			resRead.GetResult().GetFirstName(), payload.GetFirstName(),
		)
	}
	if resRead.GetResult().GetMiddleName() != payload.GetMiddleName() {
		t.Fatalf("unexpected contact middle name: have %s; expected %s",
			resRead.GetResult().GetMiddleName(), payload.GetMiddleName(),
		)
	}
	if resRead.GetResult().GetLastName() != payload.GetLastName() {
		t.Fatalf("unexpected contact last name: have %s; expected %s",
			resRead.GetResult().GetFirstName(), payload.GetLastName(),
		)
	}
	if resRead.GetResult().GetPrimaryEmail() != payload.GetPrimaryEmail() {
		t.Fatalf("unexpected contact primary email: have %s; expected %s",
			resRead.GetResult().GetPrimaryEmail(), payload.GetPrimaryEmail(),
		)
	}
}

func TestDeleteContactEntry(t *testing.T) {
	dbTest.Reset(t)
	client, close := newContactsClient(t)
	defer close()
	res, err := client.Create(DefaultContext(t), &pb.CreateContactRequest{
		Payload: &pb.Contact{
			FirstName:    "Frodo",
			LastName:     "Baggins",
			PrimaryEmail: "frodo@shire.com",
			Emails: []*pb.Email{
				{
					Address: "frodo-baggins@gondor.gov",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unable to create new contact: %s", err)
	}
	// ensure the contact was created
	if _, err := client.Read(DefaultContext(t), &pb.ReadContactRequest{
		Id: res.GetResult().GetId(),
	}); err != nil {
		t.Fatalf("unable to get contact: %s", err)
	}
	// delete the contact
	if _, err := client.Delete(DefaultContext(t), &pb.DeleteContactRequest{
		Id: res.GetResult().GetId(),
	}); err != nil {
		t.Fatalf("unable to delete contact: %s", err)
	}
	// ensure the contact was deleted
	if _, err := client.Read(DefaultContext(t), &pb.ReadContactRequest{
		Id: res.GetResult().GetId(),
	}); err == nil {
		t.Fatal("expected non-nil error when deleting empty entry")
	}
}
