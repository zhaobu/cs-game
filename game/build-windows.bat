set GOOS=windows

go build -tags consul -o bin/win/gate.exe cy/game/gate

go build -tags consul -o bin/win/center.exe cy/game/center

go build -tags consul -o bin/win/club.exe cy/game/club

REM go build -tags consul -o bin/win/ddz.exe cy/game/logic/ddz

go build -tags consul -o bin/win/changshu.exe cy/game/logic/changshu

pause