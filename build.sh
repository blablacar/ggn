#!/bin/bash
set -x
set -e
start=`date +%s`
dir=$( dirname $0 )

ENVS="linux\ndarwin"
[ -z "$1" ] || ENVS="$1"


[ -f $GOPATH/bin/godep ] || go get github.com/tools/godep
[ -f /usr/bin/upx ] || (echo "upx is required to build dgr" && exit 1)

# clean
rm -Rf $dir/dist/*-amd64

#save dep
godep save ./... || true

# format && test
gofmt -w -s .
godep go test -cover $dir/...

if [ -z ${VERSION} ]; then
    VERSION=0
fi


# build
#if `command -v parallel >/dev/null 2>&1`; then
#    echo -e "$ENVS" | parallel --will-cite -j10 --workdir . "GOOS={} GOARCH=amd64 godep go build -o dist/{}-amd64/ggn"
#else
    for e in `echo -e "$ENVS"`; do
        GOOS="$e" GOARCH=amd64 godep go build --ldflags "-s -w -X main.BuildDate=`date -u '+%Y-%m-%d_%H:%M'` \
 -X main.GgnVersion=${VERSION} \
 -X main.CommitHash=`git rev-parse HEAD`" \
    -o "dist/${e}-amd64/ggn" ${dir}/

    upx ${dir}/dist/${e}-amd64/ggn

    done
#fi

# install
cp $dir/dist/`go env GOHOSTOS`-`go env GOHOSTARCH`/ggn $GOPATH/bin/ggn

end=`date +%s`
echo "Duration : $((end-start))s"
