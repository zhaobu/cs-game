set GOOS=windows

go build -tags consul -o bin/win/http.exe game/http

go build -tags consul -o bin/win/gate.exe game/gate

go build -tags consul -o bin/win/center.exe game/center

go build -tags consul -o bin/win/club.exe game/club

REM go build -tags consul -o bin/win/ddz.exe game/logic/ddz

go build -o bin/win/initdb.exe game/script/initdb

go build -tags consul -o bin/win/changshu.exe game/logic/changshu

REM pause