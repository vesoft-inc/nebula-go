#!/bin/bash

# Script: add_apache_header.sh
# Description: Add Apache License 2.0 header to source code files
# Usage: ./add_apache_header.sh <directory_path>

target_dir=$1
if [[ ! -d "$target_dir" ]]; then
  echo "Error: Directory does not exist or not specified."
  echo "Usage: $0 <directory_path>"
  exit 1
fi

# Apache License 2.0 header (adjust COMMENT_PREFIX, year and copyright holder as needed)
LICENSE_HEADER_TEMPLATE="Licensed under the Apache License, Version 2.0 (the \"License\");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an \"AS IS\" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License."

# Process files in directory (extend FILE_EXTENSIONS as needed)
FILE_EXTENSIONS=("*.c" "*.go" "*.h" "*.cpp" "*.java" "*.py" "*.sh" "*.js" "*.html" "*.css")
for ext in "${FILE_EXTENSIONS[@]}"; do
  find "$target_dir" -type f -name "$ext" -not -path "*/proto/*" | while read -r file; do
    # Check if file already has Apache License header
    if grep -q "Licensed under the Apache License" "$file"; then
      echo "Skipping file '$file' (already contains Apache License header)"
      continue
    fi

    # Check if file contains other licenses (e.g., MIT), skip if found
    if grep -q "licensed under the MIT" "$file"; then
      echo "Skipping file '$file' (contains other license)"
      continue
    fi

    # Generate appropriate comment header based on file type
    case "$ext" in
      *.c|*.h|*.cpp|*.java)
        COMMENT_PREFIX=" * "
        HEADER="/*
 * Copyright $(date +%Y) The Author. All Rights Reserved.
 *
 * $COMMENT_PREFIX$LICENSE_HEADER_TEMPLATE
 */"
        # Convert newlines to newline followed by COMMENT_PREFIX
        HEADER=$(echo "$HEADER" | sed "s|^|$COMMENT_PREFIX|")
        HEADER="/*\n * Copyright $(date +%Y) vesoft inc. All Rights Reserved.\n *\n$HEADER\n */"
        ;;
      *.py|*.sh)
        COMMENT_PREFIX="# "
        HEADER="# Copyright $(date +%Y) vesoft inc. All Rights Reserved.\n#\n$(echo "$LICENSE_HEADER_TEMPLATE" | sed "s|^|$COMMENT_PREFIX|")"
        ;;
      *.js|*.html|*.css)
        COMMENT_PREFIX=" * "
        HEADER="/*\n * Copyright $(date +%Y) vesoft inc. All Rights Reserved.\n *\n$(echo "$LICENSE_HEADER_TEMPLATE" | sed "s|^|$COMMENT_PREFIX|")\n */"
        ;;
      *.go)
        COMMENT_PREFIX="// "
        HEADER="// Copyright $(date +%Y) vesoft inc. All Rights Reserved.\n//\n$(echo "$LICENSE_HEADER_TEMPLATE" | sed "s|^|$COMMENT_PREFIX|")"
        ;;
      *)
        COMMENT_PREFIX=""
        HEADER="$LICENSE_HEADER_TEMPLATE"
        ;;
    esac

    # Create temporary file and write license header with original content
    temp_file=$(mktemp)
    echo -e "$HEADER" > "$temp_file"
    
    # Remove leading empty lines from original file, then append content
    sed '/./,/^$/!d' "$file" >> "$temp_file" 2>/dev/null || cat "$file" >> "$temp_file"
    
    # Replace original file with temporary file
    mv "$temp_file" "$file"
    echo "Added Apache License header to '$file'"
  done
done

echo "Operation completed."