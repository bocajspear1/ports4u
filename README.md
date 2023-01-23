<h1 align="center">Ports4u</h1>

<h2 align="center">No port? No Problem</h2>
<br/>

![Ports4u Status](https://github.com/bocajspear1/ports4u/actions/workflows/test.yml/badge.svg)
![Open Issues](https://img.shields.io/github/issues-raw/bocajspear1/ports4u)
![License](https://img.shields.io/github/license/bocajspear1/ports4u)

## What is Ports4u?

Ports4u is a Golang-based application built for malware network traffic analysis, replacing something like InetSim. It detects attempted connections to ports and creates a quick listener on that port. It takes advantage of the multiple attempts TCP will take if it doesn't get back a response from a SYN packet. Ports4u utilizes `iptables` to block the RST packets that would otherwise notify of a closed port.

Ports4u also supports forwarding traffic based on the data it receives to real services it runs. For example, if it gets HTTP on another port, it forwards that traffic to the HTTP server on port 80.

Ports4u is currently oriented to be used in a Docker container.

## Building

Assumes you have Docker installed.

Run:
```
make build
```

## Supported Services

Ports4u currently runs the following services:

* HTTP on port 80
* TLS on port 443

## Data

All logs are available in the `logs` subdirectory. Ports4u will create it on startup if not already present.

### HOST-PORT.log

Contains the contents sent to Ports4u, with the remote IP and port in the filename.

Data recieved is prepended with 
```
<<<<<<<< <REMOTE_IP> ----------------------------
```

While data sent is prepended with:
```
>>>>>>>> <REMOTE_IP> ----------------------------
```
### ip_list.txt

Contains a newline separated list of IPs seen being connected to.

### domains.txt

Contains a newline separated list of domains been requested.

### conn_list.txt

Contains a list of connections seen, the format is:

```
tcp or udp|<IP>|<PORT>
```

## TODO

* More services to forward to