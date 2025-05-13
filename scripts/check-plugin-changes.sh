#!/bin/bash

echo "=== Plugin Change Detection Script ==="
echo "Current directory: $(pwd)"
echo "CIRCLE_PULL_REQUEST: ${CIRCLE_PULL_REQUEST}"
echo "Arguments: arch=$1, gotestsum=$2"
echo "RACE: ${RACE}"

# Check if we're in a pull request
if [[ ${CIRCLE_PULL_REQUEST##*/} == "" ]]; then
    echo "Not in a pull request context, running all tests"
    GOARCH=$1 ./$2 -- ${RACE} -short ./...
    exit 0
fi

# Get the list of changed files and show them
CHANGED_FILES=$(git diff origin/master --name-only)
echo "=== Changed Files ==="
echo "$CHANGED_FILES"
echo "====================="

# Initialize variables
FOUND_MATCH=false
TARGET_DIR=""
MULTIPLE_PLUGINS=false
PLUGIN_DIRS=()

# Loop through the changed files
for FILE in $CHANGED_FILES; do
    echo "Checking file: $FILE"
    # Check if the file is in any of the plugin directories and is a relevant file type
    if [[ $FILE =~ ^plugins/(aggregators|parsers|inputs|outputs)/([^/]+)/.*\.(go|mod|sum)$ ]]; then
        # Extract the plugin type and directory name
        PLUGIN_TYPE=${BASH_REMATCH[1]}
        DIR_NAME=${BASH_REMATCH[2]}

        # Set the target directory for testing
        CURRENT_TARGET="plugins/$PLUGIN_TYPE/$DIR_NAME"
        echo "Found plugin change: $CURRENT_TARGET"

        # Check if we already found a different plugin directory
        if [ "$FOUND_MATCH" = true ] && [ "$TARGET_DIR" != "$CURRENT_TARGET" ]; then
            echo "Multiple plugin directories detected"
            MULTIPLE_PLUGINS=true
            break
        fi

        TARGET_DIR=$CURRENT_TARGET
        FOUND_MATCH=true

        # Add to array for potential future use (ShellCheck SC2199 fix)
        if [[ ! " ${PLUGIN_DIRS[*]} " =~ " ${CURRENT_TARGET} " ]]; then
            PLUGIN_DIRS+=("$CURRENT_TARGET")
        fi
    fi
done

# Show what plugins were detected
if [ "$FOUND_MATCH" = true ]; then
    echo "=== Detected Plugin Directories ==="
    printf '%s\n' "${PLUGIN_DIRS[@]}"
    echo "==================================="
fi

# Run tests based on what we found
if [ "$MULTIPLE_PLUGINS" = true ]; then
    echo "Changes detected in multiple plugin directories"
    echo "Running all tests due to multiple plugin changes"
    GOARCH=$1 ./$2 -- ${RACE} -short ./...
elif [ "$FOUND_MATCH" = true ]; then
    echo "Changes detected in $TARGET_DIR"

    # Check if directory exists
    if [ ! -d "$TARGET_DIR" ]; then
        echo "Warning: $TARGET_DIR directory not found, running all tests"
        GOARCH=$1 ./$2 -- ${RACE} -short ./...
    else
        echo "Running tests only in $TARGET_DIR"
        # Change to the target directory and run tests
        (cd "$TARGET_DIR" && GOARCH=$1 ../../$2 -- ${RACE} -short ./...) || {
            echo "Error: Tests failed in $TARGET_DIR, exit code: $?"
            exit 1
        }
        echo "Tests completed successfully in $TARGET_DIR"
    fi
else
    echo "No changes detected in plugin directories, or changes in non-plugin files"
    echo "Running all tests"
    GOARCH=$1 ./$2 -- ${RACE} -short ./...
fi

echo "=== Script execution completed ==="