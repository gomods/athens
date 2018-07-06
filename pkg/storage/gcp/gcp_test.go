package gcp

import (
	"context"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/gomods/athens/pkg/config"
	athensStorage "github.com/gomods/athens/pkg/storage"
	"google.golang.org/api/option"
)

func (g *GcpTests) TestNewWithCredentials() {
	r := g.Require()
	store, err := NewWithCredentials(g.context, g.options)
	r.NoError(err)
	r.NotNil(store.bucket)
}

func (g *GcpTests) TestSaveGetListRoundTrip() {
	r := g.Require()
	store, err := NewWithCredentials(g.context, g.options)
	r.NoError(err)

	// test saving to storage
	err = store.Save(g.context, g.module, g.version, mod, zip, info)
	r.NoError(err)
	// check save was successful
	err = exists(g.context, g.options, g.bucket, g.module, g.version)
	r.NoError(err)

	// test fetching from storage
	version, err := store.Get(g.context, g.module, g.version)
	defer version.Zip.Close()
	r.NoError(err)

	r.Equal(mod, version.Mod)
	r.Equal(info, version.Info)

	gotZip, err := ioutil.ReadAll(version.Zip)
	r.NoError(version.Zip.Close())
	r.NoError(err)
	r.Equal(zip, gotZip)

	// test not found
	_, err = store.Get(g.context, "never", "there")
	notFoundErr := athensStorage.ErrVersionNotFound{Module: "never", Version: "there"}
	r.EqualError(notFoundErr, err.Error())

	// test listing modules
	// versionList, err := store.List(g.module)
	// r.NoError(err)
	// r.Equal(1, len(versionList))
	// r.Equal(g.version, versionList[0])
}

func exists(ctx context.Context, cred option.ClientOption, bucket, mod, ver string) error {
	client, err := storage.NewClient(ctx, cred)
	if err != nil {
		return err
	}
	bkt := client.Bucket(bucket)

	if _, err := bkt.Object(config.PackageVersionedName(mod, ver, "mod")).Attrs(ctx); err != nil {
		return err
	}
	if _, err := bkt.Object(config.PackageVersionedName(mod, ver, "info")).Attrs(ctx); err != nil {
		return err
	}
	if _, err := bkt.Object(config.PackageVersionedName(mod, ver, "zip")).Attrs(ctx); err != nil {
		return err
	}
	return nil
}
