go build -tags consul -o bin/linux/gate cy/game/gate

go build -tags consul -o bin/linux/center cy/game/center

go build -tags consul -o bin/linux/club cy/game/club

# go build -tags consul -o bin/linux/ddz cy/game/logic/ddz

go build -o bin/linux/initdb cy/game/script/initdb

go build -tags consul -o bin/linux/changshu cy/game/logic/changshu
