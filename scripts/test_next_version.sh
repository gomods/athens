#!/usr/bin/env bash
# Tests for scripts/next_version.sh. Runs it against throwaway git repos.
set -euo pipefail

script="$(cd "$(dirname "$0")" && pwd)/next_version.sh"
fail=0

# new_repo <initial-tag|""> : empty repo with one commit, optionally tagged.
new_repo() {
  dir="$(mktemp -d)"
  cd "$dir"
  git init -q
  git config user.email t@example.com
  git config user.name t
  git config commit.gpgsign false
  echo a > README.md && git add . && git commit -q -m "chore: init"
  [ -z "${1:-}" ] || git tag "$1"
}

# commit <subject> [file]: commit a change to <file> (default pkg/x.go).
commit() {
  local f="${2:-pkg/x.go}"
  mkdir -p "$(dirname "$f")"
  echo "$RANDOM" >> "$f"
  git add . && git commit -q -m "$1"
}

# expect <name> <expected-stdout>
expect() {
  local got
  got="$(bash "$script" 2>/dev/null || echo "ERROR")"
  if [ "$got" = "$2" ]; then
    echo "ok   - $1"
  else
    echo "FAIL - $1: expected '$2', got '$got'"
    fail=1
  fi
}

new_repo v1.2.3; commit "feat(api): add thing"
expect "feat -> minor" "v1.3.0"

new_repo v1.2.3; commit "fix: repair thing"
expect "fix -> patch" "v1.2.4"

new_repo v1.2.3; commit "perf: faster thing"
expect "perf -> patch" "v1.2.4"

new_repo v1.2.3; commit "feat!: drop thing"
expect "type! -> major" "v2.0.0"

new_repo v1.2.3; git commit -q --allow-empty -m "refactor: x" -m "BREAKING CHANGE: gone"
expect "BREAKING CHANGE footer -> major" "v2.0.0"

new_repo v1.2.3; commit "feat: a"; commit "fix: b"
expect "feat wins over fix" "v1.3.0"

new_repo v1.2.3; commit "docs: words" docs/x.md
expect "docs only -> none" ""

new_repo v1.2.3; commit "chore: tests" pkg/x_test.go
expect "test file only -> none" ""

new_repo v1.2.3; commit "chore: ci" .github/workflows/ci.yml
expect "CI only -> none" ""

new_repo v1.2.3; commit "update-go-pkg(deps): bump x" go.mod
expect "non-conventional runtime change -> patch" "v1.2.4"

new_repo v1.2.3; commit "chore(deps): bump base image" cmd/proxy/Dockerfile
expect "chore on Dockerfile -> patch" "v1.2.4"

new_repo v1.2.3
expect "no commits since tag -> none" ""

new_repo "";  commit "feat: first"
expect "no tag, feat -> v0.1.0" "v0.1.0"

new_repo v0.9.0; git tag v0.10.0; commit "fix: z"
expect "version sort (0.10.0 > 0.9.0)" "v0.10.1"

new_repo v1.2.3; git tag v1.3.0-rc1; commit "fix: z"
expect "pre-release tag is refused" "ERROR"

exit "$fail"