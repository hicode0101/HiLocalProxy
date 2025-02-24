#!/bin/sh

if [ -z "$1" ]; then
    arg1="v1.0.0"
else
    arg1=$1
fi

echo "build version: $arg1"

echo "----------------------------"

echo "building linux amd64 ..."

cd ../HiProxyServer

GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-w -s"

cd ../build

mkdir -p HiProxyServer-linux-amd64-$arg1
cp ../HiProxyServer/config.json HiProxyServer-linux-amd64-$arg1/config.json
cp ../HiProxyServer/HiProxyServer HiProxyServer-linux-amd64-$arg1/HiProxyServer

zip -r HiProxyServer-linux-amd64-$arg1.zip HiProxyServer-linux-amd64-$arg1

rm -rf HiProxyServer-linux-amd64-$arg1


echo "building win amd64 ..."

cd ../HiProxyServer

GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-w -s"

cd ../build

mkdir -p HiProxyServer-windows-amd64-$arg1
cp ../HiProxyServer/config.json HiProxyServer-windows-amd64-$arg1/config.json
cp ../HiProxyServer/HiProxyServer.exe HiProxyServer-windows-amd64-$arg1/HiProxyServer.exe

zip -r HiProxyServer-windows-amd64-$arg1.zip HiProxyServer-windows-amd64-$arg1

rm -rf HiProxyServer-windows-amd64-$arg1


echo "--------build finish--------"
