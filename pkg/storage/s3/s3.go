package s3

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/endpointcreds"
	"github.com/aws/aws-sdk-go/aws/defaults"
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
// - AWS_FORCE_PATH_STYLE	- [optional]
// For information how to get your keyId and access key turn to official aws docs: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html
type Storage struct {
	bucket   string
	uploader s3manageriface.UploaderAPI
	s3API    s3iface.S3API
	timeout  time.Duration
}

// New creates a new AWS S3 CDN saver
func New(s3Conf *config.S3Config, timeout time.Duration, options ...func(*aws.Config)) (*Storage, error) {
	const op errors.Op = "s3.New"

	awsConfig := defaults.Config()
	// remove anonymous credentials from the default config so that
	// session.NewSession can auto-resolve credentials from role, profile, env etc.
	awsConfig.Credentials = nil
	awsConfig.Region = aws.String(s3Conf.Region)
	for _, o := range options {
		o(awsConfig)
	}

	if !s3Conf.UseDefaultConfiguration {
		credProviders := defaults.CredProviders(awsConfig, defaults.Handlers())
		endpointcreds := []credentials.Provider{
			endpointcreds.NewProviderClient(*awsConfig, defaults.Handlers(), endpointFrom(s3Conf.CredentialsEndpoint, s3Conf.AwsContainerCredentialsRelativeURI)),
			&credentials.StaticProvider{
				Value: credentials.Value{
					AccessKeyID:     s3Conf.Key,
					SecretAccessKey: s3Conf.Secret,
					SessionToken:    s3Conf.Token,
				},
			},
		}

		credProviders = append(endpointcreds, credProviders...)
		awsConfig.Credentials = credentials.NewChainCredentials(credProviders)
	}

	awsConfig.S3ForcePathStyle = aws.Bool(s3Conf.ForcePathStyle)
	awsConfig.CredentialsChainVerboseErrors = aws.Bool(true)
	if s3Conf.Endpoint != "" {
		awsConfig.Endpoint = aws.String(s3Conf.Endpoint)
	}

	// create a session with creds
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, errors.E(op, err)
	}

	uploader := s3manager.NewUploader(sess)

	return &Storage{
		bucket:   s3Conf.Bucket,
		uploader: uploader,
		s3API:    uploader.S3,
		timeout:  timeout,
	}, nil
}

func endpointFrom(credentialsEndpoint string, relativeURI string) string {
	return credentialsEndpoint + relativeURI
}
