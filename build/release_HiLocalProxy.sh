#!/bin/sh

if [ -z "$1" ]; then
    arg1="v1.0.0"
else
    arg1=$1
fi

echo "build version: $arg1"

echo "----------------------------"

echo "building linux amd64 ..."

cd ../HiLocalProxy

GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-w -s"

cd ../build

mkdir -p HiLocalProxy-linux-amd64-$arg1
cp ../HiLocalProxy/config.json HiLocalProxy-linux-amd64-$arg1/config.json
cp ../HiLocalProxy/HiLocalProxy HiLocalProxy-linux-amd64-$arg1/HiLocalProxy

zip -r HiLocalProxy-linux-amd64-$arg1.zip HiLocalProxy-linux-amd64-$arg1

rm -rf HiLocalProxy-linux-amd64-$arg1


echo "building win amd64 ..."

cd ../HiLocalProxy

GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-w -s"

cd ../build

mkdir -p HiLocalProxy-windows-amd64-$arg1
cp ../HiLocalProxy/config.json HiLocalProxy-windows-amd64-$arg1/config.json
cp ../HiLocalProxy/HiLocalProxy.exe HiLocalProxy-windows-amd64-$arg1/HiLocalProxy.exe

zip -r HiLocalProxy-windows-amd64-$arg1.zip HiLocalProxy-windows-amd64-$arg1

rm -rf HiLocalProxy-windows-amd64-$arg1


echo "--------build finish--------"
