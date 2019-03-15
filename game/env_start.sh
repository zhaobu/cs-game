pkill consul
nohup consul agent -dev -client 0.0.0.0 > consul.log 2>&1 &
echo "consul start"
sudo service mongod stop &
sudo service mongod restart &
echo "mongod start"

sudo service redis stop &
sudo redis-server /etc/redis/6379.conf  &
echo "redis6379 start"
