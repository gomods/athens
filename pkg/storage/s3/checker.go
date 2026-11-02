package s3

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Exists implements the (./pkg/storage).Checker interface
// returning true if the module at version exists in storage.
func (s *Storage) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "s3.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	type checkResult struct {
		exists bool
		err    error
	}

	files := []string{"info", "mod", "zip"}
	resChan := make(chan checkResult, len(files))
	cancelingCtx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			ok, err := s.headObject(cancelingCtx, config.PackageVersionedName(module, version, file))
			resChan <- checkResult{exists: ok, err: err}
		}(file)
	}
	exists := true
	var err error
	for range files {
		res := <-resChan
		if res.err != nil {
			err = res.err
			break
		}
		if !res.exists {
			exists = false
			break
		}
	}
	cancel()
	wg.Wait()
	return exists, err
}

func (s *Storage) headObject(ctx context.Context, key string) (bool, error) {
	if s.cdnEndpoint != nil {
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, s.cdnObjectURL(key), nil)
		if err != nil {
			return false, err
		}
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return false, err
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK:
			return true, nil
		case http.StatusNotFound:
			return false, nil
		default:
			return false, fmt.Errorf("unexpected CDN status %d for HEAD %s", resp.StatusCode, key)
		}
	}

	_, err := s.s3API.HeadObject(
		ctx,
		&s3.HeadObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		var aerr smithy.APIError
		if errors.AsErr(err, &aerr) && aerr.ErrorCode() == "NotFound" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
