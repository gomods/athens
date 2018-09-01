# check_deps.ps1
# this script checks for changes to the files go.mod and go.sum
#
# this is intended to be used in your CI tests
#
# on encountering any changes for these files the script runs go mod verify
# to check for any conflicts in versions or digests
# which on any exit code > 0 would suggest that action should be taken
# before a pull request can be merged.
git remote set-branches --add origin master ; git fetch
$ChangedFiles=$(git diff --name-only origin/master)

# in the case that ChangedFiles contains go.mod or go.sum run go mod verify
$contains= $ChangedFiles | Select-String -Pattern "go.mod|go.sum"
if ($contains.length -gt 0) {
	go mod verify
}