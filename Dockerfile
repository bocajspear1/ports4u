FROM golang:1.17-alpine 

RUN apk add libpcap-dev iptables gcc musl-dev openssl

COPY . /opt/ports4u 

ENV IFACE="eth0"

RUN cd /opt/ports4u && go build -o ports4u cmd/ports4u/main.go 

RUN chmod +x /opt/ports4u/entrypoint.sh

WORKDIR /opt/ports4u/

ENTRYPOINT ["/opt/ports4u/entrypoint.sh"]