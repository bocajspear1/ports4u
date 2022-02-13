#!/bin/sh

verify_data () {
    LOG_NAME=$1
    DATA=$2
    if grep -q "${DATA}" /tmp/${LOG_NAME}; then
        echo "Found data"
    else 
        echo "Failed to find data"
        exit 1
    fi
} 

make build 

NAME=ports4u_test

docker stop ${NAME} 2>/dev/null

docker run -d --rm --cap-add=NET_ADMIN --cap-add=NET_RAW --name ${NAME} -it ports4u

GATEWAY=$(docker inspect ${NAME} | grep "Gateway\":" | head -n 1 | sed 's_^[ \t]*"Gateway": "\(.*\)".*_\1_')
IP_ADDR=$(docker inspect ${NAME} | grep "IPAddress\":" | head -n 1 | sed 's_^[ \t]*"IPAddress": "\(.*\)".*_\1_')

echo "hello how are you" | nc -q 2 ${IP_ADDR} 8080

echo ""
echo "\nChecking basic ports"

LOG_NAME=${GATEWAY}-8080.log

docker cp ${NAME}:/opt/ports4u/logs/${LOG_NAME} /tmp/${LOG_NAME}
verify_data ${LOG_NAME} "hello how are you"
rm /tmp/${LOG_NAME}

echo ""
echo "\nChecking default HTTP"

LOG_NAME=${GATEWAY}-80.log

curl -s http://${IP_ADDR} >/dev/null

docker cp ${NAME}:/opt/ports4u/logs/${LOG_NAME} /tmp/${LOG_NAME}

verify_data ${LOG_NAME} "<<<<<<<<"
verify_data ${LOG_NAME} "GET / HTTP"
rm /tmp/${LOG_NAME}

echo ""
echo "\nChecking default HTTPS"

LOG_NAME=${GATEWAY}-443.log

curl -s -k https://${IP_ADDR} >/dev/null

docker cp ${NAME}:/opt/ports4u/logs/${LOG_NAME} /tmp/${LOG_NAME}

verify_data ${LOG_NAME} "<<<<<<<<"
verify_data ${LOG_NAME} "GET / HTTP"
verify_data ${LOG_NAME} ">>>>>>>>"
verify_data ${LOG_NAME} "Session Invalid"
rm /tmp/${LOG_NAME}

echo ""
echo "Checking offport HTTP"

LOG_NAME=${GATEWAY}-9999.log

curl -s http://${IP_ADDR}:9999 >/dev/null

docker cp ${NAME}:/opt/ports4u/logs/${LOG_NAME} /tmp/${LOG_NAME}

verify_data ${LOG_NAME} "<<<<<<<<"
verify_data ${LOG_NAME} "GET / HTTP"
verify_data ${LOG_NAME} ">>>>>>>>"
verify_data ${LOG_NAME} "Session Invalid"
rm /tmp/${LOG_NAME}

echo ""
echo "Checking offport HTTPS"

LOG_NAME=${GATEWAY}-3000.log

curl -s -k https://${IP_ADDR}:3000 >/dev/null

docker cp ${NAME}:/opt/ports4u/logs/${LOG_NAME} /tmp/${LOG_NAME}

verify_data ${LOG_NAME} "<<<<<<<<"
verify_data ${LOG_NAME} "GET / HTTP"
verify_data ${LOG_NAME} ">>>>>>>>"
verify_data ${LOG_NAME} "Session Invalid"
rm /tmp/${LOG_NAME}