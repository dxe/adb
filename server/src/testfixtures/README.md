# Test fixtures

This package contains shared fixtures for tests.

Since this package imports the `model` package, using it in that package would
create a cyclical dependency. So, fixtures needed by `model` are located in
`model/activist_test_utils.go`.
