#!/bin/bash

# check_conflicts.sh
# this script checks for changes to files OTHER THAN *.go, go.mod and go.sum
# ensuring no git merge conflict artifacts are being commited.
# i.e. <<<<<<<HEAD or ======= or >>>>>>>branch-name
#
# this is intended to be used in your CI tests
#
# the script will exit with code 1 on finding any matches, causing the
# CI build to fail. Merge conflict artifacts must be removed before continuing.

git remote set-branches --add origin master && git fetch
COUNT=$(git diff -- . ':1*.go' ':!go.mod' ':!go.sum' origin/master | grep -Ec "^[\+~][<>=]{7}\w{0,}")

if (($COUNT > 0));then
  exit 1
fi
