# Test Data

This folder contains test data and test vectors copied from various repos, pull requests, or branches.

# Instructions

## Running all tests

`go test ./...`

## Running host-call tests

### Running all host-call tests

`go test ./PolkaVM -run TestHostCall`

### Running a specific test

`go test ./PolkaVM -run TestHostCall/Lookup/hostLookupNONE`

### Running multiple tests

`go test` supports using a regex (requires partial match) to select a subset of tests to run. Common examples:

To run all tests on the Info function:

`go test ./PolkaVM -run "TestHostCall/Info"`

To run tests on all general functions:

`go test ./PolkaVM -run "TestHostCall/(Lookup|Info|Read|Write)"`

To run two Info tests:

`go test ./PolkaVM -run "TestHostCall/Info/hostInfo(NONE|OK)"`

To run two tests on two separate functions:

`go test ./PolkaVM -run "TestHostCall/(Info/hostInfoNONE|Lookup/hostLookupNONE)"`
