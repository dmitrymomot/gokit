#!/bin/bash

# Script to find README files with unclosed code blocks
# If a file has an odd number of ``` tags, it likely has an unclosed code block

find /Users/dmitrymomot/Dev/gokit -name "README.md" -not -path "*/\_legacy_packages/*" | while read -r file; do
    # Count the number of ``` occurrences
    count=$(grep -c "^\`\`\`" "$file")
    
    # If count is odd, the file has unclosed code blocks
    if (( count % 2 != 0 )); then
        echo "UNBALANCED: $file ($(basename $(dirname "$file"))) - $count code blocks"
    fi
done
