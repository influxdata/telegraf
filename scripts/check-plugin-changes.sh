#!/bin/bash

echo "=== Plugin Change Detection Script ===" >&2
echo "Current directory: $(pwd)" >&2
echo "CIRCLE_PULL_REQUEST: ${CIRCLE_PULL_REQUEST}" >&2

# Default to running all tests
TEST_PATH="./..."

# Check if we're in a pull request
if [[ ${CIRCLE_PULL_REQUEST##*/} == "" ]]; then
    echo "Not in a pull request context, running all tests" >&2
    echo "$TEST_PATH"
    exit 0
fi

# Get the list of changed files and show them
CHANGED_FILES=$(git diff origin/master --name-only)
echo "=== Changed Files ===" >&2
echo "$CHANGED_FILES" >&2
echo "=====================" >&2

# Initialize variables
FOUND_MATCH=false
TARGET_DIR=""
MULTIPLE_PLUGINS=false
PLUGIN_DIRS=()

# Loop through the changed files
for FILE in $CHANGED_FILES; do
    echo "Checking file: $FILE" >&2
    # Check if the file is in any of the plugin directories and is a relevant file type
    if [[ $FILE =~ ^plugins/(aggregators|parsers|inputs|outputs)/([^/]+)/.*\.(go|mod|sum)$ ]]; then
        # Extract the plugin type and directory name
        PLUGIN_TYPE=${BASH_REMATCH[1]}
        DIR_NAME=${BASH_REMATCH[2]}

        # Set the target directory for testing
        CURRENT_TARGET="plugins/$PLUGIN_TYPE/$DIR_NAME"
        echo "Found plugin change: $CURRENT_TARGET" >&2

        # Check if we already found a different plugin directory
        if [ "$FOUND_MATCH" = true ] && [ "$TARGET_DIR" != "$CURRENT_TARGET" ]; then
            echo "Multiple plugin directories detected" >&2
            MULTIPLE_PLUGINS=true
            break
        fi

        TARGET_DIR=$CURRENT_TARGET
        FOUND_MATCH=true

        # Add to array for potential future use
        if [[ ! " ${PLUGIN_DIRS[*]} " =~ \ $CURRENT_TARGET\  ]]; then
            PLUGIN_DIRS+=("$CURRENT_TARGET")
        fi
    fi
done

# Show what plugins were detected
if [ "$FOUND_MATCH" = true ]; then
    echo "=== Detected Plugin Directories ===" >&2
    printf '%s\n' "${PLUGIN_DIRS[@]}" >&2
    echo "===================================" >&2
fi

# Determine the test path based on what we found
if [ "$MULTIPLE_PLUGINS" = true ]; then
    echo "Changes detected in multiple plugin directories" >&2
    echo "Using test path: $TEST_PATH (all tests)" >&2
elif [ "$FOUND_MATCH" = true ]; then
    # Check if directory exists
    if [ ! -d "$TARGET_DIR" ]; then
        echo "Warning: $TARGET_DIR directory not found, running all tests" >&2
        echo "Using test path: $TEST_PATH (all tests)" >&2
    else
        TEST_PATH="./$TARGET_DIR/..."
        echo "Changes detected in $TARGET_DIR" >&2
        echo "Using test path: $TEST_PATH (selective)" >&2
    fi
else
    echo "No changes detected in plugin directories, or changes in non-plugin files" >&2
    echo "Using test path: $TEST_PATH (all tests)" >&2
fi

echo "=== Script execution completed ===" >&2
echo "$TEST_PATH"