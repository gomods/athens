#!/usr/bin/env bash
# Compute the next semver tag from conventional commits since the latest vX.Y.Z tag in the
# CURRENT git repo (run it at the repo root, with full history and tags fetched).
# Prints "vX.Y.Z" to stdout when a release is warranted; prints nothing (exit 0) otherwise.
#
#   feat                          -> minor
#   fix, perf                     -> patch
#   '<type>!:' or BREAKING CHANGE -> major
#   anything else                 -> none, unless a runtime file changed since the last tag
#                                    (see below), which raises "none" to "patch"
#
# Diagnostics go to stderr so callers can capture stdout as the version.
set -euo pipefail

last="$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n1)"
if [ -n "$last" ]; then
  base="${last#v}"
  range=("${last}..HEAD")
else
  base="0.0.0"
  range=("HEAD")
fi

# The glob above also matches pre-release or four-part tags; refuse them instead of computing a
# bogus version.
if ! [[ "$base" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "::error::latest release tag '${last}' is not a plain vX.Y.Z tag; refusing to compute a bump" >&2
  exit 1
fi
IFS=. read -r major minor patch <<<"$base"

subjects="$(git log "${range[@]}" --format='%s')"
bodies="$(git log "${range[@]}" --format='%B')"

level="none"
# Breaking: a known conventional type with '!' before ':', or a BREAKING CHANGE footer.
cc_types='build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test'
if grep -Eq "^(${cc_types})(\([^)]*\))?!:" <<<"$subjects" || grep -Eq '^BREAKING[ -]CHANGE: ' <<<"$bodies"; then
  level="major"
elif grep -Eq '^feat(\([^)]*\))?:' <<<"$subjects"; then
  level="minor"
elif grep -Eq '^(fix|perf)(\([^)]*\))?:' <<<"$subjects"; then
  level="patch"
fi

# Commit type alone under-releases: Dependabot bumps to go.mod or the Dockerfile are not titled
# fix/feat, but they change the image. If nothing above was releasable and a file that affects
# the built image changed since the last tag, ship a patch. The allowlist below is narrow on
# purpose, so an unknown path errs toward releasing. Skipped before the first tag (no baseline).
if [ "$level" = "none" ] && [ -n "$last" ]; then
  non_runtime='^(\.github/|docs/|e2etests/|test/|CODEOWNERS$|\.gitignore$|LICENSE[^/]*$|\.golangci[^/]*$|[^/]*\.md$|.*_test\.go$|scripts/(next_version|test_next_version)\.sh$)'
  # --no-renames reports the old path too. Capture first: `git diff | grep -q` can SIGPIPE
  # under pipefail and read as "no match".
  changed="$(git -c core.quotePath=false diff --no-renames --name-only "${last}..HEAD")"
  runtime_files="$(grep -Ev "$non_runtime" <<<"$changed" || true)"
  if [ -n "$changed" ] && [ -n "$runtime_files" ]; then
    level="patch"
    {
      echo "next_version: no feat/fix/perf/breaking commit since ${last}, but runtime files changed -> patch:"
      head -n 10 <<<"$runtime_files" | sed 's/^/  /'
      [ "$(wc -l <<<"$runtime_files")" -le 10 ] || echo "  ..."
    } >&2
  fi
fi

case "$level" in
  major) major=$((major + 1)); minor=0; patch=0 ;;
  minor) minor=$((minor + 1)); patch=0 ;;
  patch) patch=$((patch + 1)) ;;
  none)  exit 0 ;;
esac

echo "v${major}.${minor}.${patch}"