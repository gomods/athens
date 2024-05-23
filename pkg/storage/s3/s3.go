package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/endpointcreds"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
)

// Storage implements (./pkg/storage).Backend and
// also provides a function to fetch the location of a module.
// Storage uses amazon aws go SDK which expects these env variables.
// - AWS_REGION			- region for this storage, e.g 'us-west-2'
// - AWS_ACCESS_KEY_ID		- [optional]
// - AWS_SECRET_ACCESS_KEY 	- [optional]
// - AWS_SESSION_TOKEN		- [optional]
// - AWS_FORCE_PATH_STYLE	- [optional]
// For information how to get your keyId and access key turn to official aws docs: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html.
type Storage struct {
	bucket   string
	uploader *manager.Uploader
	s3API    *s3.Client
	timeout  time.Duration
}

// New creates a new AWS S3 CDN saver.
func New(s3Conf *config.S3Config, timeout time.Duration, options ...func(*aws.Config)) (*Storage, error) {
	const op errors.Op = "s3.New"

	awsConfig, err := awscfg.LoadDefaultConfig(context.TODO(), awscfg.WithRegion(s3Conf.Region))
	if err != nil {
		return nil, errors.E(op, err)
	}

	for _, o := range options {
		o(&awsConfig)
	}

	if !s3Conf.UseDefaultConfiguration {
		// credProviders := defaults.CredProviders(awsConfig, defaults.Handlers())
		endpointCreds := []aws.CredentialsProvider{
			endpointcreds.New(endpointFrom(s3Conf.CredentialsEndpoint, s3Conf.AwsContainerCredentialsRelativeURI)),
			credentials.NewStaticCredentialsProvider(s3Conf.Key, s3Conf.Secret, s3Conf.Token),
		}

		// credProviders = append(endpointCreds, credProviders...)
		awsConfig.Credentials = newChainCredentials(endpointCreds...)
	}

	// Create a session with creds.
	sess := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = s3Conf.ForcePathStyle
		if s3Conf.Endpoint != "" {
			o.BaseEndpoint = aws.String(s3Conf.Endpoint)
		}
	})

	uploader := manager.NewUploader(sess)

	return &Storage{
		bucket:   s3Conf.Bucket,
		uploader: uploader,
		s3API:    sess,
		timeout:  timeout,
	}, nil
}

func endpointFrom(credentialsEndpoint, relativeURI string) string {
	return credentialsEndpoint + relativeURI
}

// newChainCredentials is based on old credentials.NewChainCredentials in v1.
func newChainCredentials(providers ...aws.CredentialsProvider) aws.CredentialsProvider {
	return aws.NewCredentialsCache(
		aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			var errs []error

			for _, p := range providers {
				creds, err := p.Retrieve(ctx)
				if err == nil {
					return creds, nil
				}

				errs = append(errs, err)
			}

			return aws.Credentials{}, fmt.Errorf("no valid providers in chain: %s", errs)
		}),
	)
}
