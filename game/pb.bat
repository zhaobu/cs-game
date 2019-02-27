protoc.exe --go_out=./pb/ -I./pb  ./pb/common/head.proto 
protoc.exe --go_out=./pb/ -I./pb  ./pb/common/common.proto
protoc.exe --go_out=./pb/ -I./pb -I../.. inner/inner.proto
protoc.exe --go_out=./pb/ -I./pb -I../.. center/match.proto 
protoc.exe --go_out=./pb/ -I./pb -I../.. club/club.proto
protoc.exe --go_out=./pb/ -I./pb -I../.. hall/query.proto 
protoc.exe --go_out=./pb/ -I./pb -I../.. login/login.proto --proto_path=./pb/common
protoc.exe --go_out=./pb/ -I./pb -I../.. game/game.proto
protoc.exe --go_out=./pb/ -I./pb -I../.. game/ddz/ddz.proto

pause