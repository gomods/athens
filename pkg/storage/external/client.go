package external

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

type service struct {
	url string
	c   *http.Client
}

// NewClient returns an external storage client
func NewClient(url string, c *http.Client) storage.Backend {
	if c == nil {
		c = http.DefaultClient
	}
	return &service{url, c}
}

func (s *service) List(ctx context.Context, module string) ([]string, error) {
	const op errors.Op = "external.List"
	url := s.url + "/" + module + "/@v/list"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.E(op, err)
	}
	resp, err := s.c.Do(req)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.E(op, fmt.Errorf("unexpected status code: %d", resp.StatusCode), resp.StatusCode)
	}
	list := []string{}
	scnr := bufio.NewScanner(resp.Body)
	for scnr.Scan() {
		list = append(list, scnr.Text())
	}
	if scnr.Err() != nil {
		return nil, errors.E(op, scnr.Err())
	}
	return list, nil
}

func (s *service) Info(ctx context.Context, module, vsn string) ([]byte, error) {
	const op errors.Op = "external.Info"
	url := s.url + "/" + module + "/@v/" + vsn + ".info"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.E(op, err)
	}
	resp, err := s.c.Do(req)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.E(op, fmt.Errorf("none 200 status code: %v", resp.StatusCode), resp.StatusCode)
	}
	info, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return info, nil
}

func (s *service) GoMod(ctx context.Context, module, vsn string) ([]byte, error) {
	const op errors.Op = "external.GoMod"
	url := s.url + "/" + module + "/@v/" + vsn + ".mod"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.E(op, err)
	}
	resp, err := s.c.Do(req)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.E(op, fmt.Errorf("none 200 status code: %v", resp.StatusCode), resp.StatusCode)
	}
	info, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return info, nil
}

func (s *service) Zip(ctx context.Context, module, vsn string) (io.ReadCloser, error) {
	const op errors.Op = "external.Zip"
	url := s.url + "/" + module + "/@v/" + vsn + ".zip"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.E(op, err)
	}
	resp, err := s.c.Do(req)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if resp.StatusCode != 200 {
		return nil, errors.E(op, fmt.Errorf("none 200 status code: %v", resp.StatusCode), resp.StatusCode)
	}
	return resp.Body, nil
}

func (s *service) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "external.Exists"
	_, err := s.Info(ctx, module, version)
	if err != nil {
		if errors.Is(err, errors.KindNotFound) {
			err = nil
		}
		return false, err
	}
	return true, nil
}

func (s *service) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "external.Save"
	url := s.url + "/" + module + "/@v/" + version + ".save"
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		err := upload(mw, mod, info, zip)
		pw.CloseWithError(err)
	}()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, pr)
	if err != nil {
		return errors.E(op, err)
	}
	req.Header.Add("Content-Type", mw.FormDataContentType())
	resp, err := s.c.Do(req)
	if err != nil {
		return errors.E(op, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.E(op, fmt.Errorf("unexpected status code: %v", resp.StatusCode))
	}
	return nil
}

func upload(mw *multipart.Writer, mod, info []byte, zip io.Reader) error {
	defer mw.Close()
	infoW, err := mw.CreateFormFile("mod.info", "mod.info")
	if err != nil {
		return fmt.Errorf("error creating info file: %v", err)
	}
	_, err = infoW.Write(info)
	if err != nil {
		return fmt.Errorf("error writing info file: %v", err)
	}
	modW, err := mw.CreateFormFile("mod.mod", "mod.mod")
	if err != nil {
		return fmt.Errorf("error creating mod file: %v", err)
	}
	_, err = modW.Write(info)
	if err != nil {
		return fmt.Errorf("error writing mod file: %v", err)
	}
	zipW, err := mw.CreateFormFile("mod.zip", "mod.zip")
	if err != nil {
		return fmt.Errorf("error creating zip file: %v", err)
	}
	_, err = io.Copy(zipW, zip)
	if err != nil {
		return fmt.Errorf("error writing zip file: %v", err)
	}
	return nil
}

func (s *service) Delete(ctx context.Context, module, vsn string) error {
	const op errors.Op = "external.Delete"
	return errors.E(op, "unimplemented")
}
