#! /usr/bin/env bash

testHasBuildArgsWithNoBuildArgsReturnsFalse() {
  input='{"params":{"foo":"bar"}}'
  actual=$(has_build_args "$input")
  assertFalse "$actual"
}

testHasBuildArgsWithBuildArgsReturnsTrue() {
  input='{"params":{"build_args":{"foo":"bar"}}}'
  actual=$(has_build_args "$input")
  assertTrue "$actual"
}

testHasFromFileWithNoFromFileReturnsFalse() {
  input='{"params":{"build_args":{"foo":"bar"}}}'
  actual=$(has_from_file "$input")
  assertFalse "$actual"
}

testHasFromFileWithFromFileReturnsTrue() {
  input='{"params":{"build_args":{"from_file":{"foo":"bar"}}}}'
  actual=$(has_from_file "$input")
  assertTrue "$actual"
}

oneTimeSetUp() {
  . ../../assets/resource-based-build-args.sh
}

. ./shunit2

