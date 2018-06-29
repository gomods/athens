/*
Package gcp provides a storage driver to upload module files to a google
cloud platform storage bucket.

Configuration

Environment variables:

	ATHENS_GCP_BUCKET_NAME // required

Example:

	Bash:
		export ATHENS_GCP_BUCKET_NAME="fancy-pony-33928.appspot.com"
	Fish:
		set -x ATHENS_GCP_BUCKET_NAME fancy-pony-339288.appspot.com

*/
package gcp
