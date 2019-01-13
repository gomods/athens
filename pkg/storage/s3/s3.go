package s3

import (
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
)

// Storage implements (./pkg/storage).Backend and
// also provides a function to fetch the location of a module
// Storage uses amazon aws go SDK which expects these env variables
// - AWS_REGION			- region for this storage, e.g 'us-west-2'
// - AWS_ACCESS_KEY_ID		- [optional]
// - AWS_SECRET_ACCESS_KEY 	- [optional]
// - AWS_SESSION_TOKEN		- [optional]
// For information how to get your keyId and access key turn to official aws docs: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html
type Storage struct {
	bucket   string
	baseURI  *url.URL
	uploader s3manageriface.UploaderAPI
	s3API    s3iface.S3API
	timeout  time.Duration
}

// New creates a new AWS S3 CDN saver
func New(s3Conf *config.S3Config, timeout time.Duration, options ...func(*aws.Config)) (*Storage, error) {
	const op errors.Op = "s3.New"
	u, err := url.Parse(fmt.Sprintf("https://%s.s3.amazonaws.com", s3Conf.Bucket))
	if err != nil {
		return nil, errors.E(op, err)
	}

	creds := buildAWSCredentials(s3Conf)

	awsConfig := &aws.Config{
		Credentials: creds,
		Region:      aws.String(s3Conf.Region),
	}

	for _, o := range options {
		o(awsConfig)
	}

	// create a session
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, errors.E(op, err)
	}
	uploader := s3manager.NewUploader(sess)

	return &Storage{
		bucket:   s3Conf.Bucket,
		uploader: uploader,
		s3API:    uploader.S3,
		baseURI:  u,
		timeout:  timeout,
	}, nil
}

// buildAWSCredentials builds the credentials required to create a new AWS session.  It will prefer the access key ID and
// secret access key if specified in the S3Config.  Otherwise it will look for credentials in the filesystem.
func buildAWSCredentials(s3Conf *config.S3Config) *credentials.Credentials {
	if !s3Conf.UseAmbientCredentials && s3Conf.Key != "" && s3Conf.Secret != "" {
		return credentials.NewStaticCredentials(s3Conf.Key, s3Conf.Secret, s3Conf.Token)
	}

	return nil
}
