package gcp

import (
	"context"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

func (g *GcpTests) TestNewWithCredentials() {
	r := g.Require()
	store, err := NewWithCredentials(g.context, g.options)
	r.NoError(err)
	r.NotNil(store.bucket)
}

func (g *GcpTests) TestSaveGetRoundTrip() {
	r := g.Require()
	store, err := NewWithCredentials(g.context, g.options)
	r.NoError(err)

	err = store.Save(g.context, g.module, g.version, mod, zip, info)
	r.NoError(err)
	err = exists(g.context, g.options, g.bucket, g.module, g.version)
	r.NoError(err)

	version, err := store.Get(g.context, g.module, g.version)
	defer version.Zip.Close()
	r.NoError(err)

	r.Equal(mod, version.Mod)
	r.Equal(info, version.Info)

	gotZip, err := ioutil.ReadAll(version.Zip)
	r.NoError(version.Zip.Close())
	r.NoError(err)
	r.Equal(zip, gotZip)
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
