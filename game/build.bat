set GOOS=windows

go build -tags consul -o bin/gate.exe cy/game/gate

go build -tags consul -o bin/center.exe cy/game/center

go build -tags consul -o bin/club.exe cy/game/club

go build -tags consul -o bin/ddz.exe cy/game/logic/ddz

pause