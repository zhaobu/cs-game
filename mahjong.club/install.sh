#!/usr/bin/env bash

cd "$GOPATH/src/mahjong.club"
echo `pwd`

echo "go install"
go install

echo 'install finished'

cd -
