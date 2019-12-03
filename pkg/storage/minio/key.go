package minio

import (
	"net/url"
	"strings"
)

func extractKey(objectKey string) (key string, module string, version string) {
	var err error
	key, err = url.PathUnescape(objectKey)
	if err != nil {
		key = objectKey
	}

	parts := strings.Split(key, "/")
	version = parts[len(parts)-2]
	module = strings.Replace(key, version, "", -2)
	module = strings.Replace(module, "//.info", "", -1)

	return key, module, version
}
