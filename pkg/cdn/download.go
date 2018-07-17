package cdn

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/config/env"
	"github.com/gomods/athens/pkg/storage"
)

// ModVerDownloader downloads a module version from a URL
type ModVerDownloader func(ctx context.Context, baseURL, module, version string) (*storage.Version, error)

// Download downloads the module/version from url. Returns a storage.Version
// representing the downloaded module/version or a non-nil error if something
// went wrong
func Download(ctx context.Context, baseURL, module, version string) (*storage.Version, error) {
	tctx, cancel := context.WithTimeout(ctx, env.Timeout())
	defer cancel()
	getReq := func(ext string) (*http.Request, error) {
		return getRequest(tctx, baseURL, module, version, ext)
	}

	infoReq, err := getReq(".info")
	if err != nil {
		return nil, err
	}
	modReq, err := getReq(".mod")
	if err != nil {
		return nil, err
	}
	zipReq, err := getReq(".zip")
	if err != nil {
		return nil, err
	}

	info, err := getResBytes(infoReq)
	if err != nil {
		return nil, err
	}
	mod, err := getResBytes(modReq)
	if err != nil {
		return nil, err
	}
	zipRes, err := http.DefaultClient.Do(zipReq)
	if err != nil {
		return nil, err
	}

	ver := storage.Version{
		Info: info,
		Mod:  mod,
		Zip:  zipRes.Body,
	}
	return &ver, nil
}

func getResBytes(req *http.Request) ([]byte, error) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func getRequest(ctx context.Context, baseURL, module, version, ext string) (*http.Request, error) {
	u, err := join(baseURL, module, version, ext)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return req, nil
}

func join(baseURL string, module, version, ext string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	packageVersionedName := config.PackageVersionedName(module, version, ext)
	u.Path = path.Join(u.Path, packageVersionedName)
	return fmt.Sprint(u), nil
}
