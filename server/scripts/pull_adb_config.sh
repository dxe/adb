#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $DIR/..

if [[ -d adb-config ]] ; then
    cd adb-config
    git pull
else
    git clone 'git@github.com:dxe/adb-config'
fi
