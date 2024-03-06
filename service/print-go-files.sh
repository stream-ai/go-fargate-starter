#!/bin/bash

# List all Go files in the repository
go_files=$(git ls-tree HEAD --name-only -r | grep '\.go$')

# Loop through each Go file and print its contents
for file in $go_files; do
    echo "==> Contents of $file <=="
    cat "$file"
    echo # Print a newline for better readability between files
done

