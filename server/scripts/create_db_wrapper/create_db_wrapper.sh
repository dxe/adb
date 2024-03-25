#!/bin/bash

set -e

# if DXE_DEV_EMAIL is not set, prompt user for their email
if [ -z ${DXE_DEV_EMAIL+x} ]; then
    echo "Please enter your development email or set DXE_DEV_EMAIL to never see this message again: "
    read DXE_DEV_EMAIL
fi
go run ./create_db.go --dev-email="${DXE_DEV_EMAIL}"
