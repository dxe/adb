#!/bin/bash

set -e

# if DXE_DEV_EMAIL is not set, prompt user for their email
if [ -z ${DXE_DEV_EMAIL+x} ]; then
    echo "Please enter your development email: "
    read DXE_DEV_EMAIL
fi
go run ./scripts/create_db.go --dev-email="${DXE_DEV_EMAIL}"
