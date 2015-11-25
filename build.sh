#!/bin/bash
set -e
start=`date +%s`
dir=$( dirname $0 )

ENVS="darwin\nlinux"
[ -z "$1" ] || ENVS="$1"


[ -f $GOPATH/bin/godep ] || go get github.com/tools/godep

# clean
rm -Rf $dir/dist/*-amd64

# binary
#[ -f $GOPATH/bin/go-bindata ] || go get -u github.com/jteeuwen/go-bindata/...
#mkdir -p $dir/dist/bindata

#[ -f $dir/aci-bats/aci-bats.aci ] || $dir/aci-bats/build.sh
#cp $dir/aci-bats/aci-bats.aci $dir/dist/bindata
#[ -f $dir/dist/bindata/busybox ] || cp /bin/busybox $dir/dist/bindata/
#[ -f $dir/dist/bindata/attributes-merger ] || wget "https://github.com/blablacar/attributes-merger/releases/download/0.1/attributes-merger" -O $dir/dist/bindata/attributes-merger
#[ -f $dir/dist/bindata/confd ] || wget "https://github.com/kelseyhightower/confd/releases/download/v0.10.0/confd-0.10.0-linux-amd64" -O $dir/dist/bindata/confd
#go-bindata -nomemcopy -pkg dist -o $dir/dist/bindata.go $dir/dist/bindata/...

# format && test
gofmt -w -s .
godep go test -cover $dir/...

# build
if `command -v parallel >/dev/null 2>&1`; then
    echo -e "$ENVS" | parallel --will-cite -j10 --workdir . "GOOS={} GOARCH=amd64 godep go build -o dist/{}-amd64/ggn"
#    mv dist/windows-amd64/ggn dist/windows-amd64/ggn.exe
else
    for e in `echo -e "$ENVS"`; do
        GOOS="$e" GOARCH=amd64 godep go build -o "dist/${e}-amd64/ggn"
    done
fi

cp $dir/dist/linux-amd64/ggn $GOPATH/bin/ggn

end=`date +%s`
echo "Duration : $((end-start))s"
