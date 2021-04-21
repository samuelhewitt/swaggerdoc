#!/bin/bash
set -eo pipefail

VERSION=1.5.0

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
    echo "  -l, --skip-linters   Skip linters"
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

# Remove old swaggerdoc binary, forcing a rebuild, as we don't want to
# accidentally add the zip archive to it multiple times
/bin/rm -f $OUTPUT_PATH

# Determine the git commit
if which git 2>&1 > /dev/null; then
  if [ -z "`git status --porcelain`" ]; then
    STATE=clean
  else
    STATE=dirty
  fi
  GIT_VERSION=`git rev-parse HEAD`-$STATE
else
  GIT_VERSION=Unknown
fi

# Build the binary
LINK_FLAGS="-X github.com/richardwilkes/toolbox/cmdline.AppVersion=$VERSION"
LINK_FLAGS="$LINK_FLAGS -X github.com/richardwilkes/toolbox/cmdline.GitVersion=$GIT_VERSION"
go build -v -ldflags=all="$LINK_FLAGS" -o "$OUTPUT_PATH" .

echo "Created $OUTPUT_PATH"