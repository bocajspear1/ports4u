package services

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

var iptablesPath string = ""

func getIPtablesPath() string {
	if iptablesPath == "" {
		cmd := exec.Command("which", "iptables")

		var out bytes.Buffer
		cmd.Stdout = &out

		err := cmd.Run()

		if err != nil {
			log.Fatal(err)
		}

		iptablesPath = strings.TrimSpace(out.String())
		log.Printf("Got iptables path at %s\n", iptablesPath)
	}
	return iptablesPath
}

func AllowTCPPort(port uint16) {
	iptables := getIPtablesPath()
	cmd := exec.Command(iptables, "-I", "OUTPUT", "1", "-p", "tcp", "--dport", fmt.Sprintf("%d", port), "-j", "ACCEPT")

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()

	if err != nil {
		log.Println("out:", outb.String(), "err:", errb.String())
		log.Fatal(err)
	}
}

func BlockTCPPort(port uint16) {
	iptables := getIPtablesPath()
	cmd := exec.Command(iptables, "-D", "OUTPUT", "-p", "tcp", "--dport", fmt.Sprintf("%d", port), "-j", "ACCEPT")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()

	if err != nil {
		log.Println("out:", outb.String(), "err:", errb.String())
		log.Fatal(err)
	}
}
