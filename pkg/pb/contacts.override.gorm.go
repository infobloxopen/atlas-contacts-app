package pb

import (
	"fmt"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BeforeToORM will add the primary e-mail to the list of e-mails if it isn't
// present already
func (m *Contact) BeforeToORM(ctx context.Context, c *ContactORM) error {
	if m.PrimaryEmail != "" {
		for _, mail := range m.Emails {
			if mail.Address == m.PrimaryEmail {
				return nil
			}
		}
		c.Emails = append(c.Emails, &EmailORM{Address: m.PrimaryEmail, IsPrimary: true})
	}
	return nil
}

// AfterToPB copies the primary e-mail address from the DB to the special PB field
func (m *ContactORM) AfterToPB(ctx context.Context, c *Contact) error {
	if len(m.Emails) == 0 {
		return nil
	}
	// find the primary e-mail in list of e-mails from DB
	for _, addr := range m.Emails {
		if addr != nil && addr.IsPrimary {
			c.PrimaryEmail = addr.Address
			break
		}
	}
	return nil
}

// Hooks and direct db access:
// Hooks Documentation: http://gorm.io/docs/hooks.html
// Database access: http://gorm.io/docs/sql_builder.html
// Below is an example that shows how to add a hook that runs before the Update method and how to access the database directly.

// BeforeUpdate runs before the Update method and checks to see if the ID provided already exists.
func (m *ContactORM) BeforeUpdate(scope *gorm.Scope) (err error) {
	// Create a new db access point
	dbc := scope.NewDB()

	// To execute commands directly :
	// dbc.Exec("SOME SQL COMMAND")

	// Executes the raw sql cmd and gets query response in rows
	rows, err := dbc.Raw("SELECT * FROM contacts WHERE id=" + fmt.Sprint(m.Id)).Rows() // (*sql.Rows, error)
	defer rows.Close()
	if err != nil {
		return err
	}
	// Check to see if a row with id already exists
	if rows.Next() {
		// Contact exists proceed to default update method
		return nil
	} else {
		return fmt.Errorf("NOT_EXISTS")
	}
}

// Overriding CRUD Methods:
// For the example below we will be overriding the Read method.
// To override a CRUD method we need to find the Custom method (CustomCreate, CustomRead, CustomUpdate, CustomDelete)
// from the gorm file (contacts.pb.gorm.go) and implement it below.

// CustomRead method overrides the default Read function and adds custom errors with multiple details.
func (m *ContactsDefaultServer) CustomRead(ctx context.Context, req *ReadContactRequest) (*ReadContactResponse, error) {
	res, err := DefaultReadContact(ctx, &Contact{Id: req.GetId()}, m.DB)
	if err != nil {
		st := status.Newf(codes.Internal, "Unable to read contact. Error %v", err)
		st, _ = st.WithDetails(errdetails.New(codes.Internal, "CustomRead", "Custom error message"))
		st, _ = st.WithDetails(errdetails.New(codes.Internal, "CustomRead", "Another custom error message"))
		return nil, st.Err()
	}
	return &ReadContactResponse{Result: res}, nil
}
