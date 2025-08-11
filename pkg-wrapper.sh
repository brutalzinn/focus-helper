#!/bin/bash

# This script intercepts calls to pkg-config and replaces the
# request for the old webkit2gtk-4.0 with the newer 4.1.

new_args=()
for arg in "$@"; do
  if [ "$arg" == "webkit2gtk-4.0" ]; then
    # When we see the old version, substitute the new one.
    new_args+=("webkit2gtk-4.1")
  else
    # Otherwise, keep the argument as is.
    new_args+=("$arg")
  fi
done

# Execute the real pkg-config with the corrected arguments.
exec /usr/bin/pkg-config "${new_args[@]}"