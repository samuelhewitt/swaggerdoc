#!/bin/bash
set -eo pipefail

EXE=${GOPATH}/bin/swaggerdoc

# Remove old swaggerdoc binary, forcing a rebuild, as we don't want to
# accidentally add the zip archive to it multiple times
/bin/rm -f ${EXE}

# Build the binary
go install -v .

# Add the zip file system to the binary
cd dist
zip --quiet -9 ../dist.zip *
cd ..
cat dist.zip >> ${EXE}
zip --quiet -A ${EXE}
/bin/rm -f dist.zip
