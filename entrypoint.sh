#!/bin/sh

iptables -A OUTPUT -p icmp -j DROP
iptables -A OUTPUT -p tcp --tcp-flags RST RST -j DROP
/opt/ports4u/ports4u