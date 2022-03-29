#!/bin/sh

iptables -A OUTPUT -p icmp --icmp-type destination-unreachable -j DROP
iptables -A OUTPUT -p tcp --tcp-flags RST RST -j DROP
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 365 -nodes -subj "/C=US/ST=New York/L=New York/O=Global Security/OU=IT Department/CN=limit.com"
/opt/ports4u/ports4u