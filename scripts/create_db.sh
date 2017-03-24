#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

rm "$DIR/../adb.db"
echo '.q' | sqlite -init "$DIR/create_sqlite_db.sql" "$DIR/../adb.db"
