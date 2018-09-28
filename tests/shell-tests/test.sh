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

testOnlyFileContentsReadIntoKVP() {
  tempfile1=$(mktemp)
  tempfile2=$(mktemp)
  echo "qux" >> "$tempfile1"
  echo "eggs" >> "$tempfile2"
  full_input='{"params":{"build_args":{"from_file":{"foo":"'
  full_input+="$tempfile1"
  full_input+='","spam":"'
  full_input+="$tempfile2"
  full_input+='"}}}}'
  build_args=$(buildArgExtractionCopiedFromProd "$full_input")
  expected='{"spam":"eggs","foo":"qux"}'
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

