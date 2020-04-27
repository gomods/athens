package cmd

import (
	"context"
	"io"

	"github.com/gomods/athens/pkg/storage"
)

// returns the info bytes, mod, and zip (in that order)
func getCompleteMod(
	ctx context.Context,
	getter storage.Getter,
	mod string,
	ver string,
) ([]byte, []byte, io.ReadCloser, error) {
	infoBytes, err := getter.Info(ctx, mod, ver)
	if err != nil {
		return nil, nil, nil, err
	}
	modBytes, err := getter.GoMod(ctx, mod, ver)
	if err != nil {
		return nil, nil, nil, err
	}
	zip, err := getter.Zip(ctx, mod, ver)
	if err != nil {
		return nil, nil, nil, err
	}
	return infoBytes, modBytes, zip, nil
}

func transfer(
	ctx context.Context,
	cataloger storage.Cataloger,
	from storage.Getter,
	to storage.Saver,
) error {
	token := ""
	// TODO: parallelize this
	for {
		pathParams, newToken, err := cataloger.Catalog(ctx, token, 100)
		if err != nil {
			return err
		}
		for _, pathParam := range pathParams {
			mod := pathParam.Module
			ver := pathParam.Version
			infoBytes, modBytes, zip, err := getCompleteMod(
				ctx,
				from,
				mod,
				ver,
			)
			if err != nil {
				return err
			}
			saveErr := to.Save(
				ctx,
				mod,
				ver,
				modBytes,
				zip,
				infoBytes,
			)
			if saveErr != nil {
				return err
			}
		}
		if newToken == "" {
			return nil
		}
		token = newToken
	}

	return nil
}
