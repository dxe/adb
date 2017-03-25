#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

rm -f "$DIR/../adb.db"
echo '.q' | sqlite3 -init "$DIR/create_sqlite_db.sql" "$DIR/../adb.db"
