package svc

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"encoding/base64"
	"strings"

	"strconv"

	"fmt"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/errors"
	"github.com/infobloxopen/atlas-app-toolkit/query"
	"github.com/infobloxopen/atlas-contacts-app/pkg/pb"
	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/status"
)

// NewProfilesServer returns an instance of the default profiles server interface
func NewProfilesServer(database *gorm.DB) (pb.ProfilesServer, error) {
	return &profilesServer{&pb.ProfilesDefaultServer{DB: database}}, nil
}

type profilesServer struct {
	*pb.ProfilesDefaultServer
}

// NewGroupsServer returns an instance of the default groups server interface
func NewGroupsServer(database *gorm.DB) (pb.GroupsServer, error) {
	return &groupsServer{&pb.GroupsDefaultServer{DB: database}}, nil
}

type groupsServer struct {
	*pb.GroupsDefaultServer
}

// NewContactsServer returns an instance of the default contacts server interface
func NewContactsServer(database *gorm.DB) (pb.ContactsServer, error) {
	return &contactsServer{&pb.ContactsDefaultServer{DB: database}}, nil
}

type contactsServer struct {
	*pb.ContactsDefaultServer
}

// List wraps default ContactsDefaultServer.List implementation by adding
// application specific page token implementation.
// Actually the service supports "composite" pagination in a specific way:
// - limit and offset are still supported but without page token
// - if an user requests page token and provides limit then limit value will be
//	 used as a step for all further requests
//		page_toke = null & limit = 2 -> page_token=base64(offset=3:limit=2)
// - if an user requests page token and provides offset then only first time
//	 the provided offset is applied
//		page_token = null & offset = 2 & limit = 2 -> page_token=base64(offset=2+2:limit=2)
func (s *contactsServer) List(ctx context.Context, in *empty.Empty) (*pb.ListContactsResponse, error) {
	page, err := gateway.Pagination(ctx)
	if err != nil {
		return nil, err
	}

	ptoken := page.GetPageToken()
	// do not handle page token
	if ptoken == "" {
		return s.ContactsDefaultServer.List(ctx, in)
	}

	// decode provided token (null means a client is requesting new token)
	// update context with new pagination request
	if ptoken != "null" {
		page.Offset, page.Limit, err = DecodePageToken(ptoken)
		if err != nil {
			return nil, err
		}
		ctx = gateway.NewPaginationContext(ctx, page)
	}

	// forward request to default implementation
	resp, err := s.ContactsDefaultServer.List(ctx, in)
	if err != nil {
		return nil, err
	}

	// prepare and set response page info
	var pinfo query.PageInfo
	if length := len(resp.Results); length == 0 {
		pinfo.SetLastToken()
	} else {
		pinfo.PageToken = EncodePageToken(page.GetOffset()+int32(length), page.DefaultLimit())
	}
	if err := gateway.SetPageInfo(ctx, &pinfo); err != nil {
		return nil, err
	}

	return resp, nil
}

// DecodePageToken decodes page token from the user's request.
// Return error if provided token is malformed or contains ivalid values,
// otherwise return offset, limit.
func DecodePageToken(ptoken string) (offset, limit int32, err error) {
	errC := errors.InitContainer()
	data, err := base64.StdEncoding.DecodeString(ptoken)
	if err != nil {
		return 0, 0, errC.New(codes.InvalidArgument, "invalid page token - %s", err)
	}
	vals := strings.SplitN(string(data), ":", 2)
	if len(vals) != 2 {
		return 0, 0, errC.New(codes.InvalidArgument, "malformed page token")
	}

	o, err := strconv.Atoi(vals[0])
	if err != nil {
		errC.Set("page_token", codes.InvalidArgument, "invalid offset value %s", err)
		errC.WithField("offset", "incorrect offset value")
	}

	l, err := strconv.Atoi(vals[1])
	if err != nil {
		errC.Set("page_token", codes.InvalidArgument, "invalid limit value %s", err)
		errC.WithField("limit", "incorrect limit value")
	}

	limit = int32(l)
	offset = int32(o)

	if err := errC.IfSet(codes.InvalidArgument, "invalid page token"); err != nil {
		return 0, 0, errC
	}

	return limit, offset, nil
}

// EncodePageToken encodes offset and limit to a string in application specific
// format (offset:limit) in base64 encoding.
func EncodePageToken(offset, limit int32) string {
	data := fmt.Sprintf("%d:%d", offset, limit)
	return base64.StdEncoding.EncodeToString([]byte(data))
}
