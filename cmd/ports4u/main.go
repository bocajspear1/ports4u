package main

import (
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/bocajspear1/ports4u/internal/services"
	"github.com/bocajspear1/ports4u/internal/watcher"
)

func main() {
	iface := os.Getenv("IFACE")
	ifaceAddr := ""
	ifaceMAC := ""

	ok := false
	counter := 0
	for !ok && counter < 4 {
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
				ip, _, err := net.ParseCIDR(addrs[0].String())
				ifaceAddr = ip.String()
				ifaceMAC = strings.ToLower(localIface.HardwareAddr.String())
			}
		}

		if ifaceAddr == "" {
			counter += 1
			time.Sleep(500 * time.Millisecond)
		} else {
			ok = true
		}
	}

	if !ok {
		log.Fatalf("Unable to get address for interface %s, does it exist?\n", iface)
	} else {
		log.Printf("Got address of %s for iface %s\n", ifaceAddr, iface)
	}

	services.AddRedirect(ifaceAddr, iface)

	ignorePorts := []uint16{80, 443}
	watcher.StartWatcher(iface, ifaceAddr, ifaceMAC, ignorePorts)

	httpService := services.NewHTTPService()
	go httpService.Start(ifaceAddr, 80)

	tlsService := services.NewTLSService()
	go tlsService.Start(ifaceAddr, 443)

	dnsService := services.NewDNSService()
	dnsService.Start(ifaceAddr, 53)
}
