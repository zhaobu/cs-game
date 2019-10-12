set GOOS=linux

go build -tags consul -o bin/linux/http game/http

go build -tags consul -o bin/linux/gate game/gate

go build -tags consul -o bin/linux/center game/center

go build -tags consul -o bin/linux/club game/club

REM go build -tags consul -o bin/linux/ddz game/logic/ddz

go build -o bin/linux/initdb game/script/initdb

go build -tags consul -o bin/linux/changshu game/logic/changshu

REM pause