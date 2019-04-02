# ./clear_redis.sh
# sleep 2

cd bin/linux

./club >/dev/null &
echo "club start success" &

./center >/dev/null &
echo "center start success" &

./gate -addr 192.168.1.128:9876 >/dev/null &
echo "gate start success" &

# ./ddz >/dev/null &
# echo "ddz start success" 
