/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package watcher

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/bocajspear1/ports4u/internal/services"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// https://godoc.org/github.com/google/gopacket
// https://godoc.org/github.com/google/gopacket/pcap

var serviceMap map[uint16]*services.PortService

func logPacket(protocol gopacket.LayerType, port uint16, outFile *os.File) {

	if port == 0 {
		return
	}

}

func watcherRun(ipaddr string, iface string, pcapFilter string) {

	serviceMap = make(map[uint16]*services.PortService)

	pcapHandle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)

	if err != nil {
		log.Fatalf("Could not open interface %s for listening: %s\n", iface, err)
		return
	}

	err = pcapHandle.SetBPFFilter(pcapFilter)

	if err != nil {
		log.Fatalf("Could not set filter %s for interface %s: %s\n", pcapFilter, iface, err)
		return
	}

	log.Println("Missed port watcher listening...")

	var tcp layers.TCP
	var udp layers.UDP
	var eth layers.Ethernet
	var ip4 layers.IPv4
	var ip6 layers.IPv6

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &udp)
	decodedPacket := []gopacket.LayerType{}

	pcapPacket := gopacket.NewPacketSource(pcapHandle, pcapHandle.LinkType())
	for packet := range pcapPacket.Packets() {

		err = parser.DecodeLayers(packet.Data(), &decodedPacket)
		// if err != nil {
		// 	fmt.Printf("err: %s\n", err)
		// 	continue
		// }

		for _, layerType := range decodedPacket {
			if layerType == layers.LayerTypeTCP {
				if tcp.SYN {
					log.Printf("Got TCP on port %d\n", tcp.DstPort)
					port := uint16(tcp.DstPort)
					_, ok := serviceMap[port]
					if !ok {
						log.Printf("New port service for %d\n", tcp.DstPort)
						serviceMap[port] = services.NewPortService()
					} else {
						if !serviceMap[port].IsActive() {
							delete(serviceMap, port)
							serviceMap[port] = services.NewPortService()
						}
					}

					if !serviceMap[port].IsActive() {
						log.Printf("Starting\n")
						go serviceMap[port].Start(ipaddr, port)
					} else {
						log.Printf("Using existing\n")
					}

				}
			} else if layerType == layers.LayerTypeUDP {

			}
		}

	}

}

// StartWatcher starts the missed port watcher
func StartWatcher(iface string, ignorePorts []uint16) {

	newFilter := ""

	// Get interface address info so we don't intercept data sent by us
	ifaceInfo, err := net.InterfaceByName(iface)
	if err != nil {
		log.Fatalf("Could not get interface %s\n", iface)
		return
	}

	addrs, err := ifaceInfo.Addrs()
	if err != nil {
		log.Fatalf("Could not get interface %s addresses\n", iface)
		return
	}

	ifaceAddr := ""

	portFilter := ""
	for _, port := range ignorePorts {
		filterPart := "not port " + fmt.Sprintf("%d", port)
		if portFilter != "" {
			portFilter += " and " + filterPart
		} else {
			portFilter += filterPart
		}
	}

	newFilter += portFilter + " "

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if newFilter != "" {
			newFilter += " and not src host " + ip.String()
		} else {
			ifaceAddr = ip.String()
			newFilter += " not src host " + ip.String()
		}

	}

	log.Printf("Watcher filter: \n\n%s\n\n", newFilter)

	log.Println("Starting missed port watcher...")
	go watcherRun(ifaceAddr, iface, newFilter)
}
