#!/bin/bash
set -eo pipefail

VERSION=1.0.1

PRIMARY_GOPATH=`echo $GOPATH | sed -e 's/:.*//'`
if [ -z $PRIMARY_GOPATH ]; then
    PRIMARY_GOPATH=`echo $GOPATH | sed -e 's/.*://'`
fi

EXE=$GOPATH/bin/swaggerdoc

# Remove old swaggerdoc binary, forcing a rebuild, as we don't want to
# accidentally add the zip archive to it multiple times
/bin/rm -f $EXE

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
go install -v -ldflags=all="$LINK_FLAGS" .

# Add the zip file system to the binary
cd dist
zip --quiet -9 ../dist.zip *
cd ..
cat dist.zip >> $EXE
zip --quiet -A $EXE
/bin/rm -f dist.zip
