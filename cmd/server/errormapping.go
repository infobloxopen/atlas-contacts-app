package main

import (
	"github.com/infobloxopen/atlas-app-toolkit/errors"

	"google.golang.org/grpc/codes"

	"context"
)

var ErrorMappings = []errors.MapFunc{
	errors.NewMapping(
		errors.CondEq("NOT_EXISTS"),
		errors.NewContainer(
			codes.InvalidArgument, "The resource does not exist.",
		).WithField(
			"path/id", "the specified object was not found."),
	),
	errors.NewMapping(
		// Here CondAnd without condition functions serves as 'default'.
		errors.CondAnd(),
		errors.MapFunc(func (ctx context.Context, err error) (error, bool) {
			return errors.NewContainer(codes.Internal, "Error: %s", err), true
		}),
	),

}
