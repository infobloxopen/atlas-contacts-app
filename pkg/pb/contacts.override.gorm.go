package pb

import (
	"strings"

	"github.com/infobloxopen/atlas-app-toolkit/auth"
	"github.com/infobloxopen/atlas-app-toolkit/query"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BeforeToORM will add the primary e-mail to the list of e-mails if it isn't
// present already
func (m *Contact) BeforeToORM(ctx context.Context, c *ContactORM) error {
	emails := []*Email{}
	if m.PrimaryEmail != "" {
		for _, mail := range m.Emails {
			if mail.Address != m.PrimaryEmail {
				emails = append(emails, mail)
			}
		}

		m.Emails = emails

		accountID, err := auth.GetAccountID(ctx, nil)
		if err != nil {
			return err
		}

		c.Emails = []*EmailORM{&EmailORM{Address: m.PrimaryEmail, AccountID: accountID, IsPrimary: true}}
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

// BeforeUpdate method deals with synthetic field primary_email. If primary_email is set and no emails given,
// it reads all emails, filter out current primary email and appends to FieldMask emails address.
func (m *UpdateContactRequest) BeforeUpdate(ctx context.Context, req *UpdateContactRequest, db *gorm.DB) (context.Context, *gorm.DB, error) {
	var hasPrimaryEmail, hasEmails bool
	for _, p := range req.GetFields().GetPaths() {
		if p == "PrimaryEmail" {
			hasPrimaryEmail = true
		}
		if p == "Emails" {
			hasEmails = true
		}
	}

	if hasPrimaryEmail && !hasEmails {
		db := db.Preload("Emails")
		res, err := DefaultReadContact(ctx, &Contact{Id: req.GetPayload().GetId()}, db)
		if err != nil {
			code := codes.Internal
			if err == gorm.ErrRecordNotFound {
				code = codes.NotFound
			}
			st := status.Newf(code, "Unable to read contact. Error %v", err)
			return ctx, db, st.Err()
		}

		emails := []*Email{}
		for _, e := range res.Emails {
			if e.Address != res.PrimaryEmail {
				emails = append(emails, e)
			}
		}
		m.Payload.Emails = emails
		m.Fields.Paths = append(m.Fields.Paths, "Emails")
	}

	return ctx, db, nil

}

// Overriding CRUD Methods:
// For the example below we will be overriding the Read method.
// To override a CRUD method we need to find the Custom method (CustomCreate, CustomRead, CustomUpdate, CustomDelete)
// from the gorm file (contacts.pb.gorm.go) and implement it below.

// CustomRead method overrides the default Read function and adds custom errors with multiple details.
func (m *ContactsDefaultServer) CustomRead(ctx context.Context, req *ReadContactRequest) (*ReadContactResponse, error) {
	res, err := DefaultReadContact(ctx, &Contact{Id: req.GetId()}, m.DB)
	if err != nil {
		code := codes.Internal
		if err == gorm.ErrRecordNotFound {
			code = codes.NotFound
		}
		st := status.Newf(code, "Unable to read contact. Error %v", err)
		st, _ = st.WithDetails(errdetails.New(codes.InvalidArgument, "CustomRead", "Example of custom error message"))
		st, _ = st.WithDetails(errdetails.New(codes.InvalidArgument, "CustomRead", "Another example of custom error message"))
		return nil, st.Err()
	}
	return &ReadContactResponse{Result: res}, nil
}

type FilteringIteratorCallback func(path []string, f interface{}) (interface{}, string)

// IterateFiltering call callback function for each condtion struct of *Filtering.
// Callback results override original Filtering condition and append list of joins,
// so finally IterateFiltering returns modified *Filtering with
// list of join conditions for supporting new *Filtering
func IterateFiltering(f *query.Filtering, callback FilteringIteratorCallback) []string {
	joins := []string{}

	var getOperator func(interface{}) interface{}

	doCallback := func(path []string, f interface{}) interface{} {
		res, join := callback(path, f)
		if res != nil {
			if join != "" {
				joins = append(joins, join)
			}
			return res
		}
		return f
	}

	getOperator = func(f interface{}) interface{} {
		val := f.(*query.LogicalOperator)

		left := val.GetLeft()
		switch leftVal := left.(type) {
		case *query.LogicalOperator_LeftOperator:
			val.SetLeft(getOperator(leftVal.LeftOperator))

		case *query.LogicalOperator_LeftStringCondition:
			val.SetLeft(doCallback(leftVal.LeftStringCondition.GetFieldPath(), leftVal.LeftStringCondition))

		case *query.LogicalOperator_LeftNumberCondition:
			val.SetLeft(doCallback(leftVal.LeftNumberCondition.GetFieldPath(), leftVal.LeftNumberCondition))

		case *query.LogicalOperator_LeftNullCondition:
			val.SetLeft(doCallback(leftVal.LeftNullCondition.GetFieldPath(), leftVal.LeftNullCondition))
		}

		right := val.GetRight()
		switch rightVal := right.(type) {
		case *query.LogicalOperator_RightOperator:
			val.SetRight(getOperator(rightVal.RightOperator))

		case *query.LogicalOperator_RightStringCondition:
			val.SetRight(doCallback(rightVal.RightStringCondition.GetFieldPath(), rightVal.RightStringCondition))

		case *query.LogicalOperator_RightNumberCondition:
			val.SetRight(doCallback(rightVal.RightNumberCondition.GetFieldPath(), rightVal.RightNumberCondition))

		case *query.LogicalOperator_RightNullCondition:
			val.SetRight(doCallback(rightVal.RightNullCondition.GetFieldPath(), rightVal.RightNullCondition))
		}
		return val
	}

	root := f.GetRoot()
	switch val := root.(type) {
	case *query.Filtering_Operator:
		f.SetRoot(getOperator(val.Operator))

	case *query.Filtering_StringCondition:
		f.SetRoot(doCallback(val.StringCondition.GetFieldPath(), val.StringCondition))

	case *query.Filtering_NumberCondition:
		f.SetRoot(doCallback(val.NumberCondition.GetFieldPath(), val.NumberCondition))

	case *query.Filtering_NullCondition:
		f.SetRoot(doCallback(val.NullCondition.GetFieldPath(), val.NullCondition))
	}
	return joins
}

// CustomList method overrides the default Read function and modify Filtering to support synthetic fields.
func (m *ContactsDefaultServer) CustomList(ctx context.Context, in *ListContactRequest) (*ListContactsResponse, error) {
	db := m.DB
	f := in.GetFilter()
	if f != nil {
		joins := IterateFiltering(f, supportSynteticFields())
		for _, join := range joins {
			db = db.Joins(join)
		}
	}
	res, err := DefaultListContact(ctx, db, in)
	if err != nil {
		return nil, err
	}
	return &ListContactsResponse{Results: res}, nil
}

// callback function for IterateFiltering to support "primary_email" (synthetic field) filtering
func supportSynteticFields() FilteringIteratorCallback {
	syntheticFound := false
	return func(path []string, f interface{}) (interface{}, string) {
		join := ""
		switch strings.Join(path, ".") {
		case "primary_email":
			sc, ok := f.(*query.StringCondition)
			if ok {
				sc.FieldPath = []string{"synthetic_emails", "address"}
				if !syntheticFound {
					join = "join emails synthetic_emails on contacts.id = synthetic_emails.contact_id and synthetic_emails.is_primary = true"
					syntheticFound = true
				}
				return sc, join
			}
		}
		return nil, ""
	}
}
