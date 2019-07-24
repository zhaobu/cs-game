pkill consul &
nohup ./consul agent -dev -client 0.0.0.0 > consul.log 2>&1 &
echo "consul start"

pkill mongod &
nohup ./mongod -f mongod.conf 2>&1 &
echo "mongod start"

pkill redis &
nohup ./redis-server redis-6380.conf 2>&1 &
echo "redis6379 start"

