set GOOS=windows

go build -tags consul -o bin/win/logic.exe cy/other/im/logic

go build -tags consul -o bin/win/gate.exe cy/other/im/gate

go build -tags consul -o bin/win/friend.exe cy/other/im/friend

REM pause