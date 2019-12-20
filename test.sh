#!/bin/sh

set -e

generate_cover_data() {
    for pkg in "$@"; do
      if [[ "$pkg" =~ "utility" || "$pkg" =~ "example" ]]
        then
          continue
      fi
      go test "$pkg" -cover
    done
}

generate_cover_data $(go list ./...)

