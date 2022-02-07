## What is Ports4u?

Ports4u is a Golang-based application built for malware network traffic analysis. It detects ports something is attempting to connect on and creates a quick listener on that port. It takes advantage of the multiple attempts TCP will take if it doesn't get back a response from a SYN packet. Ports4u utilizes `iptables` to block the RST packets that would otherwise notify of a closed port.

Ports4u also supports forwarding traffic based on the data it receives to real services it runs. For example, if it gets HTTP on another port, it forwards that traffic to the HTTP server on port 80.

Ports4u is currently oriented to be used in a Docker container.

## TODO

* Logging
* More services to forward to
* Build instructions