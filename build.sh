export GOOS=linux GOARCH=amd64 CGO_ENABLED=0

go build  -tags consul -o bin/linux/http game/http

go build  -tags consul -o bin/linux/gate game/gate

go build  -tags consul -o bin/linux/center game/center

go build  -tags consul -o bin/linux/club game/club

go build  -o bin/linux/initdb game/script/initdb

go build  -tags consul -o bin/linux/changshu game/logic/changshu