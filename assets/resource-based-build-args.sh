#! usr/bin/env bash

has_build_args() {
  result=$(echo "$1" | jq 'map_values(has("build_args")).params')
  if [[ "$result" == "true" ]]; then echo "0"; else echo "1"; fi
}

has_from_file() {
  result=$(echo "$1" | jq '.params | map_values(has("from_file")).build_args')
  if [[ "$result" == "true" ]]; then echo "0"; else echo "1"; fi
}

