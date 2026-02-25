#!/bin/sh

flags=""
if [[ -d adb-config ]]; then
  . adb-config/env
  flags="-prod"
fi

./adb-server $flags
