package minio

import (
	"context"
	"testing"

	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage/minio/mocks"
	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/require"
)

func TestCatalog(t *testing.T) {
	token := "test-token"
	bucketName := "test-bucket"
	pageSize := 10

	mockMinioCore := mocks.NewMinioCore(t)
	mockMinioCore.On("ListObjectsV2", bucketName, "", token, false, "", 0, "").Return(minio.ListBucketV2Result{
		Contents: []minio.ObjectInfo{
			{
				Key: "test-version/test.info",
			},
		},
	}, nil)

	st := &storageImpl{
		minioCore:  mockMinioCore,
		bucketName: bucketName,
	}

	gotAllPathsParams, gotToken, gotErr := st.Catalog(context.Background(), token, pageSize)

	expectedToken := ""
	expectedAllPathsParams := []paths.AllPathParams{{Module: "/test.info", Version: "test-version"}}

	require.NoError(t, gotErr)
	require.Equal(t, expectedToken, gotToken)
	require.Equal(t, expectedAllPathsParams, gotAllPathsParams)
}
