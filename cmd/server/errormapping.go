package main

import (
	"github.com/infobloxopen/atlas-app-toolkit/errors"

	"google.golang.org/grpc/codes"
)

var ErrorMappings = []errors.MapFunc{
	errors.NewMapping(
		errors.CondEq("NOT_EXISTS"),
		errors.NewContainer(
			codes.InvalidArgument, "The resource does not exist.",
		).WithField(
			"path/id", "the specified object was not found."),
	),
}
