#!/bin/bash
set -x
set -e
start=`date +%s`
dir=$( dirname $0 )

ENVS="darwin\nlinux"
[ -z "$1" ] || ENVS="$1"


[ -f $GOPATH/bin/godep ] || go get github.com/tools/godep

# clean
rm -Rf $dir/dist/*-amd64

# format && test
gofmt -w -s .
godep go test -cover $dir/...

# build
if `command -v parallel >/dev/null 2>&1`; then
    echo -e "$ENVS" | parallel --will-cite -j10 --workdir . "GOOS={} GOARCH=amd64 godep go build -o dist/{}-amd64/ggn"
else
    for e in `echo -e "$ENVS"`; do
        GOOS="$e" GOARCH=amd64 godep go build -o "dist/${e}-amd64/ggn"
    done
fi

godep go install

end=`date +%s`
echo "Duration : $((end-start))s"
