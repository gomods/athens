package minio

import (
	"net/url"
	"strings"
)

func extractKey(objectKey string) (key, module, version string) {
	var err error
	key, err = url.PathUnescape(objectKey)
	if err != nil {
		key = objectKey
	}

	parts := strings.Split(key, "/")
	version = parts[len(parts)-2]
	module = strings.ReplaceAll(key, version, "")
	module = strings.ReplaceAll(module, "//.info", "")

	return key, module, version
}
