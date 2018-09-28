#! /usr/bin/env bash

testHasFromFileWithNoFromFileReturnsFalse() {
  full_input='{"params":{"build_args":{"foo":"bar"}}}'
  build_args=$(buildArgExtractionCopiedFromProd "$full_input")
  actual=$(has_from_file "$build_args")
  assertFalse "$actual"
}

testHasFromFileWithFromFileReturnsTrue() {
  full_input='{"params":{"build_args":{"from_file":{"foo":"bar"}}}}'
  build_args=$(buildArgExtractionCopiedFromProd "$full_input")
  actual=$(has_from_file "$build_args")
  assertTrue "$actual"
}

testFromSingleFileContentsReadIntoKVP() {
  tempfile=$(mktemp)
  echo "qux" >> "$tempfile"
  full_input='{"params":{"build_args":{"from_file":{"foo":"'
  full_input+="$tempfile"
  full_input+='"}}}}'
  build_args=$(buildArgExtractionCopiedFromProd "$full_input")
  expected='{"foo":"qux"}'
  actual=$(elevate_from_file_kvps "$build_args")
  assertEquals "$expected" "$actual"
}

oneTimeSetUp() {
  . ../../assets/resource-based-build-args.sh
}

buildArgExtractionCopiedFromProd() {
  # https://github.com/concourse/docker-image-resource/blob/master/assets/out#L66
  result=$(echo "$1" | jq -r '.params.build_args // {}')
  echo "$result"
}

. ./shunit2

