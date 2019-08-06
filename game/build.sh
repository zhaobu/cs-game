set CGO_ENABLED=0

go build -tags netgo -tags consul -o bin/linux/http cy/game/http

GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -tags netgo -tags consul -o bin/linux/gate cy/game/gate

go build -tags netgo -tags consul -o bin/linux/center cy/game/center

go build -tags netgo -tags consul -o bin/linux/club cy/game/club

go build -tags netgo -o bin/linux/initdb cy/game/script/initdb

go build -tags netgo -tags consul -o bin/linux/changshu cy/game/logic/changshu