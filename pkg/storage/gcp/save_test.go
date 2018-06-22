package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	"google.golang.org/appengine/aetest"
)

func (g *GcpTests) TestNewStorage() {
	r := g.Require()
	ctx, done, err := aetest.NewContext()
	defer done()
	r.NoError(err)
	store, err := New(ctx, g.bucket, g.options)
	r.NoError(err)
	r.NotNil(store.bucket)
}

func (g *GcpTests) TestSave() {
	r := g.Require()
	ctx, done, err := aetest.NewContext()
	defer done()
	r.NoError(err)
	store, err := New(ctx, g.bucket, g.options)
	r.NoError(err)
	err = store.Save(ctx, g.module, g.version, mod, info, zip)
	r.NoError(err)

	err = exists(ctx, g.options, g.bucket, g.module, g.version)
	r.NoError(err)
}

func exists(ctx context.Context, cred option.ClientOption, bucket, mod, ver string) error {
	client, err := storage.NewClient(ctx, cred)
	if err != nil {
		return err
	}
	bkt := client.Bucket(bucket)

	if _, err := bkt.Object(fmt.Sprintf("%s/@v/%s.%s", mod, ver, "mod")).Attrs(ctx); err != nil {
		return err
	}
	if _, err := bkt.Object(fmt.Sprintf("%s/@v/%s.%s", mod, ver, "info")).Attrs(ctx); err != nil {
		return err
	}
	if _, err := bkt.Object(fmt.Sprintf("%s/@v/%s.%s", mod, ver, "zip")).Attrs(ctx); err != nil {
		return err
	}
	return nil
}
