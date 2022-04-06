/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package watcher

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/bocajspear1/ports4u/internal/services"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// https://godoc.org/github.com/google/gopacket
// https://godoc.org/github.com/google/gopacket/pcap

var serviceMap map[uint16]*services.PortService
var arpMap map[string]uint16

func logPacket(protocol gopacket.LayerType, port uint16, outFile *os.File) {

	if port == 0 {
		return
	}

}

func logUnique(outFilePath string, new_item string) {
	services.CheckLogDir()
	outFile, err := ioutil.ReadFile(outFilePath)
	found := false
	if err == nil {
		ips := strings.Split(string(outFile), "\n")
		for _, ip := range ips {
			if new_item == ip {
				found = true
			}
		}
	}

	if !found {
		// Log IP if it hasn't been found already
		out, err := os.OpenFile(outFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal("Failed to open " + outFilePath)
		}
		_, err = out.Write([]byte(new_item + "\n"))
		if err != nil {
			log.Fatal("Failed to write to " + outFilePath)
		}
		out.Close()
	}
}

func logUniqueAddr(new_ip string) {

	ipFilePath := "./logs/ip_list.txt"
	logUnique(ipFilePath, new_ip)

}

func logUniqueConn(new_conn string) {

	connFilePath := "./logs/conn_list.txt"
	logUnique(connFilePath, new_conn)

}

func watcherRun(ipaddr string, mac string, iface string, pcapFilter string, ignorePorts []uint16) {

	serviceMap = make(map[uint16]*services.PortService)
	arpMap = make(map[string]uint16)

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
	var arp layers.ARP

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &udp, &arp)
	decodedPacket := []gopacket.LayerType{}

	pcapPacket := gopacket.NewPacketSource(pcapHandle, pcapHandle.LinkType())
	for packet := range pcapPacket.Packets() {

		err = parser.DecodeLayers(packet.Data(), &decodedPacket)

		for _, layerType := range decodedPacket {
			if layerType == layers.LayerTypeTCP {
				// Check for SYN-only packets (connection start)
				if tcp.SYN && !tcp.ACK {

					port := uint16(tcp.DstPort)
					found := false

					// Ignore any connection created by our own container
					// This should not normally happen
					if strings.ToLower(eth.SrcMAC.String()) == mac {
						found = true
					}

					// Ignore anything going local, such as port forwarding
					if ip4.DstIP.String() == "127.0.0.1" || ip4.SrcIP.String() == "127.0.0.1" {
						found = true
					}

					logUniqueConn(fmt.Sprintf("tcp:%s:%d", ip4.DstIP.String(), port))

					// Ensure we ignore ports we provide and don't try to
					// create a listener on that port
					// We don't use a filter so we can pick up remote IPs, even if they
					// go to the ports we ignore
					for _, p := range ignorePorts {
						if p == port {
							found = true
						}
					}

					srcPort := uint16(tcp.SrcPort)
					for _, p := range ignorePorts {
						if p == srcPort {
							found = true
						}
					}

					// Create listener if it doesn't match anything above
					if !found {
						log.Printf("Got TCP on port %d\n", tcp.DstPort)
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

				}
			} else if layerType == layers.LayerTypeUDP {

				port := uint16(udp.DstPort)

				if port != 53 {
					logUniqueConn(fmt.Sprintf("udp:%s:%d", ip4.DstIP.String(), port))
				}

			} else if layerType == layers.LayerTypeIPv4 {
				if ip4.DstIP.String() != ipaddr && ip4.SrcIP.String() != ipaddr && strings.ToLower(eth.SrcMAC.String()) != mac {
					logUniqueAddr(ip4.DstIP.String())
				}

			} else if layerType == layers.LayerTypeARP {
				destIP := net.IP(arp.DstProtAddress).String()

				if ipaddr != destIP {
					// Record the local IP attempted
					logUniqueAddr(destIP)
					// We randomly might or might not respond to ARP
					// We can do this since the iptables rule will redirect any connections
					randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
					num := randGen.Intn(6)

					if num < 2 {
						_, ok := arpMap[destIP]
						if !ok {
							log.Printf("Sending ARP response for %s\n", destIP)
							localMac, err := net.ParseMAC(mac)
							if err != nil {
								log.Println(err)
							}
							newEth := layers.Ethernet{
								SrcMAC:       localMac,
								DstMAC:       eth.SrcMAC,
								EthernetType: layers.EthernetTypeARP,
							}
							newArp := layers.ARP{
								AddrType:          layers.LinkTypeEthernet,
								Protocol:          layers.EthernetTypeIPv4,
								HwAddressSize:     6,
								ProtAddressSize:   4,
								Operation:         layers.ARPReply,
								SourceHwAddress:   []byte(localMac),
								SourceProtAddress: arp.DstProtAddress,
								DstHwAddress:      []byte(arp.SourceHwAddress),
								DstProtAddress:    []byte(arp.SourceProtAddress),
							}

							newBuf := gopacket.NewSerializeBuffer()
							opts := gopacket.SerializeOptions{
								FixLengths:       true,
								ComputeChecksums: true,
							}

							err = gopacket.SerializeLayers(newBuf, opts, &newEth, &newArp)
							if err != nil {
								log.Println(err)
							}

							if err := pcapHandle.WritePacketData(newBuf.Bytes()); err != nil {
								log.Println(err)
							}
						} else {
							// We did not send before, don't send now even though it passed
							log.Printf("Found countdown, no ARP response for %s\n", destIP)
							arpMap[destIP] -= 1
							if arpMap[destIP] == 0 {
								delete(arpMap, destIP)
							}
						}

					} else {
						_, ok := arpMap[destIP]
						if !ok {
							arpMap[destIP] = 2
						} else {
							arpMap[destIP] -= 1
						}
						log.Printf("No ARP response for %s\n", destIP)
						if arpMap[destIP] == 0 {
							delete(arpMap, destIP)
						}
					}
				}

			}

		}

	}

}

// StartWatcher starts the missed port watcher
func StartWatcher(iface string, ipaddr string, mac string, ignorePorts []uint16) {

	newFilter := ""

	if newFilter != "" {
		newFilter += " and not src host " + ipaddr
	} else {
		newFilter += "not src host " + ipaddr
	}

	log.Printf("Watcher filter: \n\n%s\n\n", newFilter)

	log.Println("Starting missed port watcher...")
	go watcherRun(ipaddr, mac, iface, newFilter, ignorePorts)
}
