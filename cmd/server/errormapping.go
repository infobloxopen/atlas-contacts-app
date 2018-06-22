package main

import (
	"github.com/infobloxopen/atlas-app-toolkit/errors"

	"google.golang.org/grpc/codes"

	"context"
	"regexp"
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
		errors.CondHasPrefix("pq:"),
		errors.MapFunc(func (ctx context.Context, err error) (error, bool) {
			if res := regexp.MustCompile(`column "(\w+)" does not exist`).FindStringSubmatch(err.Error()); len(res) > 0 {
				return errors.NewContainer(codes.InvalidArgument, "Invalid collection operator parameter %q.", res[1]), true
			}

			return nil, false
		}),
	),

	errors.NewMapping(
		// Here CondAnd without condition functions serves as 'default'.
		errors.CondAnd(),
		errors.MapFunc(func (ctx context.Context, err error) (error, bool) {
			return errors.NewContainer(codes.Internal, "Error: %s", err), true
		}),
	),

}
