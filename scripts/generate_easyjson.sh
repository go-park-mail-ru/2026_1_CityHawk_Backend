#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

export GOCACHE="${GOCACHE:-$ROOT_DIR/.cache/go-build}"
export GOTMPDIR="${GOTMPDIR:-$ROOT_DIR/.cache/gotmp}"
mkdir -p "$GOCACHE" "$GOTMPDIR"

EASYJSON=(go run github.com/mailru/easyjson/easyjson)

request_files=(
  internal/auth/delivery/http/request.go
  internal/organizer/delivery/http/request.go
  internal/social/delivery/http/request.go
  internal/support/delivery/http/request.go
  internal/user/delivery/http/request.go
)

response_files=(
  internal/auth/delivery/http/response.go
  internal/organizer/delivery/http/response.go
  internal/place/delivery/http/response.go
  internal/social/delivery/http/response.go
  internal/support/delivery/http/response.go
  internal/user/delivery/http/response.go
)

for file in "${request_files[@]}"; do
  "${EASYJSON[@]}" -all -disallow_unknown_fields "$file"
done

# place request.go has a hand-written optionalString.UnmarshalJSON for nullable
# PATCH semantics, so generated stdlib MarshalJSON/UnmarshalJSON methods would
# conflict. The easyjson methods are still generated for explicit use.
"${EASYJSON[@]}" -all -disallow_unknown_fields -no_std_marshalers internal/place/delivery/http/request.go

for file in "${response_files[@]}"; do
  "${EASYJSON[@]}" -all "$file"
done
