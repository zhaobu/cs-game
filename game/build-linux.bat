set GOOS=linux

cd gate
go build -tags consul
cd ..

cd center 
go build -tags consul
cd ..

cd club 
go build -tags consul
cd ..

cd logic/ddz
go build -tags consul
cd ../..

pause