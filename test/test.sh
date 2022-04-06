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

echo ""
echo "Checking DNS"

LOG_NAME=dns-test

dig @${IP_ADDR} another.com 
dig @${IP_ADDR} test.com > /tmp/${LOG_NAME}
sed -i 's_\t_ _g' /tmp/${LOG_NAME}
verify_data ${LOG_NAME} "test.com."
verify_data ${LOG_NAME} "IN A ${IP_ADDR}"
verify_data ${LOG_NAME} ";; ANSWER SECTION:"
rm /tmp/${LOG_NAME}

LOG_NAME=domains.txt
docker cp ${NAME}:/opt/ports4u/logs/${LOG_NAME} /tmp/${LOG_NAME}
verify_data ${LOG_NAME} "test.com."
verify_data ${LOG_NAME} "another.com."
rm /tmp/${LOG_NAME}


echo ""
echo "Checking redirection"

sudo ip route add 192.168.55.0/24 via ${IP_ADDR}
TEST_IP="192.168.55.4"

LOG_NAME=${GATEWAY}-4545.log

curl -s http://${TEST_IP}:4545 >/dev/null

docker cp ${NAME}:/opt/ports4u/logs/${LOG_NAME} /tmp/${LOG_NAME}

verify_data ${LOG_NAME} "<<<<<<<<"
verify_data ${LOG_NAME} "GET / HTTP"
verify_data ${LOG_NAME} ">>>>>>>>"
verify_data ${LOG_NAME} "Session Invalid"
rm /tmp/${LOG_NAME}

echo ""
echo "Checking empty connection (no data)"

LOG_NAME=${GATEWAY}-7070.log

nc -z ${TEST_IP} 7070 

docker cp ${NAME}:/opt/ports4u/logs/${LOG_NAME} /tmp/${LOG_NAME}
verify_data ${LOG_NAME} "<<<<<<<< ${GATEWAY}"
rm /tmp/${LOG_NAME}


echo ""
echo "Testing basic UDP"

echo "hi" | nc -w 1 -u ${TEST_IP} 1111 

echo ""
echo "Testing conn_list"

LOG_NAME=conn_list.txt

docker cp ${NAME}:/opt/ports4u/logs/${LOG_NAME} /tmp/${LOG_NAME}
verify_data ${LOG_NAME} "tcp:${TEST_IP}:80"
verify_data ${LOG_NAME} "tcp:${TEST_IP}:8080"
verify_data ${LOG_NAME} "tcp:${TEST_IP}:9999"
verify_data ${LOG_NAME} "tcp:${TEST_IP}:3000"
verify_data ${LOG_NAME} "tcp:${TEST_IP}:443"
verify_data ${LOG_NAME} "tcp:${TEST_IP}:4545"
verify_data ${LOG_NAME} "tcp:${TEST_IP}:7070"
verify_data ${LOG_NAME} "udp:${TEST_IP}:1111"

rm /tmp/${LOG_NAME}