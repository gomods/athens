/*
Package gcp provides a storage driver to upload module files to a google
cloud platform storage bucket.

Configuration

Environment variables:

	ATHENS_STORAGE_GCP_BUCKET // required

Example:

	Bash:
		export ATHENS_STORAGE_GCP_BUCKET="fancy-pony-33928.appspot.com"
	Fish:
		set -x ATHENS_STORAGE_GCP_BUCKET fancy-pony-339288.appspot.com

*/
package gcp
