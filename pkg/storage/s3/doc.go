/*
Package s3 provides a storage driver to upload module files to
amazon s3 storage bucket.

Configuration

Environment variables:
    AWS_REGION 					// region for this storage, e.g 'us-west-2'
    AWS_ACCESS_KEY_ID
    AWS_SECRET_ACCESS_KEY
    AWS_SESSION_TOKEN			// [optional]
	ATHENS_S3_BUCKET_NAME
	ATHENS_S3_ENDPOINT_ADDRESS	// [optional]
	ATHENS_S3_DISABLE_SSL		// [optional]
	ATHENS_S3_FORCE_PATH_STYLE	// [optional]

For information how to get your keyId and access key turn to official aws docs: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html

Example:

	Bash:
		export AWS_REGION="us-west-2"
	Fish:
		set -x AWS_REGION us-west-2

*/
package s3
