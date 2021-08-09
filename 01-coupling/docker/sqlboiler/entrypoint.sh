#!/bin/bash
/wait-for-mysql

for dir in /src/*; do
    echo "Running in $dir"
    sqlboiler mysql --config "$dir/sqlboiler.toml" --no-tests --output "$dir/models" --pkgname models
done
