#!/bin/bash

# Get to the root directory of the Git repository
cd "$(git rev-parse --show-toplevel)"

# List all Go files in the repository
go_files=$(git ls-tree HEAD --name-only -r | grep '\.go$')

# Loop through each Go file and print its contents
while IFS= read -r file; do
    echo "==> Contents of $file <=="
    echo "\`\`\`(golang)"
    cat "$file" || echo "Error: Could not read file $file"
    echo "\`\`\`"
    echo # Print a newline for better readability between files
done <<< "$go_files"
