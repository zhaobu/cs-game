#!/usr/bin/env sh

echo "useage:./start.sh gate true"
echo "参数个数$#"

for i in "$@"; do
    echo $i
done

# 第二个参数表示是否重新build
if [ "$2" = "true" ]; then
    echo 'rebuild true'
    cd $GOPATH/src/cy/game
    if [ "$NodeType" = "game" ]; then
        CGO_ENABLED=0 GOOS=linux go build -tags consul -o bin/linux/$NodeName cy/game/logic/$NodeName
    else
        CGO_ENABLED=0 GOOS=linux go build -tags consul -o bin/linux/$NodeName cy/game/$NodeName
    fi
else
    echo 'rebuild false'
fi
cd $GOPATH/src/cy/game/bin/linux
./$NodeName

