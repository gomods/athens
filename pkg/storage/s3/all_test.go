package s3

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TestMock struct {
	APIMock
	db map[string][]byte
}

type S3Tests struct {
	suite.Suite
	client *TestMock
}

func (d *S3Tests) SetupTest() {
	d.client = getS3Mock()
}

func getS3Mock() *TestMock {
	svc := new(TestMock)

	svc.On("PutObjectWithContext", mock.AnythingOfType("*s3.PutObjectInput")).Return(func(input *s3.PutObjectInput) {
		b, e := ioutil.ReadAll(input.Body)
		if e != nil {
			log.Fatal(e)
		}

		svc.db[*input.Key] = b
	})

	return svc
}

// Verify returns error if S3 state differs from expected one
func (t *TestMock) Verify(value map[string][]byte) error {
	expectedLength := len(value)
	actualLength := len(t.db)
	if len(value) != len(t.db) {
		return fmt.Errorf("Length does not match. Expected: %d. Actual: %d", expectedLength, actualLength)
	}

	for k, v := range value {
		actual, ok := t.db[k]
		if !ok {
			return fmt.Errorf("Missing element %s", k)
		}

		if !sliceEqualCheck(v, actual) {
			return fmt.Errorf("Value for key %s does not match. Expected: %v, Actual: %v", k, v, actual)
		}
	}

	return nil
}

func sliceEqualCheck(a, b []byte) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
