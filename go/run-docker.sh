#!/bin/bash

set -euo pipefail

FILE_PATH="$1"
OUTPUT_PATH="$2"
IMAGE_NAME="wap-tool"

if [[ ! -f "$FILE_PATH" ]]; then
    echo "Error: file '$FILE_PATH' does not exist on host."
    exit 1
fi
# Absolute path of the input file
ABS_PATH=$(realpath "$FILE_PATH")
ABS_OUTPUT_PATH=$(realpath "$OUTPUT_PATH")
OUTPUT_DIR=$(dirname "$ABS_OUTPUT_PATH")
OUTPUT_FILENAME=$(basename "$ABS_OUTPUT_PATH")
if [[ ! -d "$OUTPUT_DIR" ]]; then
    echo "Error: directory $OUTPUT_DIR does not exist on host."
    exit 1
fi


docker run --rm \
  --cap-drop=ALL \
  --user "$(id -u):$(id -g)" \
  -v "$ABS_PATH:/data/input:ro" \
  -v "$OUTPUT_DIR:/output" \
  "$IMAGE_NAME" \
  /data/input \
  -o /output/"$OUTPUT_FILENAME"
