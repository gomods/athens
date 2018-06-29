package gcp_test

import (
	"context"

	"github.com/gomods/athens/pkg/storage/gcp"
	"google.golang.org/api/option"
)

func ExampleNew() {
	// Create some credentials, in this case non authenticated which is
	// suitable for testing. For more information on available client options
	// see https://godoc.org/google.golang.org/api/option#ClientOption
	opts := option.WithoutAuthentication()

	// Get a context from the background
	ctx := context.Background()

	// Get our new storage thingy
	store, err := gcp.New(ctx, opts)
	if err != nil {
		//
	}

	// do some storagey things
}
