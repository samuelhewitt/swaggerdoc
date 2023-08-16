#!/bin/bash
set -eo pipefail

if [ -z $GOPATH ]; then
  GOPATH="$HOME/go"
fi
DEFAULT_OUTPUT_PATH="$GOPATH/bin/swaggerdoc"

# Process args
while [[ "$#" -gt 0 ]]; do case $1 in
  -o|--output)
    if [ -z $2 ]; then
      echo "The $1 option requires a path argument"
      exit 1
    fi
    OUTPUT_PATH="$2"
    shift
    ;;
  -h|--help)
    echo "$0 [options]"
    echo "  -o, --output <path>  The output binary (defaults to \"$DEFAULT_OUTPUT_PATH\")"
    echo "  -h, --help           This help text"
    exit 0
    ;;
  *)
    echo "Invalid argument: $1"
    BAIL=1
    ;;
esac; shift; done

if [ -n "$BAIL" ]; then
    exit 1
fi

if [ -z $OUTPUT_PATH ]; then
  OUTPUT_PATH="$DEFAULT_OUTPUT_PATH"
fi

# Remove old swaggerdoc binary, forcing a rebuild
/bin/rm -f $OUTPUT_PATH

# Build the binary
go build -v -o "$OUTPUT_PATH" ./v2

echo "Created $OUTPUT_PATH"