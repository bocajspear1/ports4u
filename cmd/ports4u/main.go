package main

import (
	"log"
	"net"
	"os"

	"github.com/bocajspear1/ports4u/internal/services"
	"github.com/bocajspear1/ports4u/internal/watcher"
)

func main() {
	iface := os.Getenv("IFACE")
	ifaceAddr := ""

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("Unable to get interfaces\n")
	}
	for _, localIface := range ifaces {
		if localIface.Name == iface {
			addrs, err := localIface.Addrs()
			if err != nil {
				log.Fatalf("Unable to get addresses for interface %s\n", iface)
			}
			ifaceAddr = addrs[0].String()
		}
	}

	if iface == "" {
		log.Fatalf("Unable to get address for interface %s, does it exist?\n", iface)
	}

	ignorePorts := []uint16{80, 443}
	watcher.StartWatcher(iface, ignorePorts)

	httpService := services.NewHTTPService()
	go httpService.Start(ifaceAddr, 80)

	tlsService := services.NewTLSService()
	go tlsService.Start(ifaceAddr, 443)

	dnsService := services.NewDNSService()
	dnsService.Start(ifaceAddr, 53)
}
