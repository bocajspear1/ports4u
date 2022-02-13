package services

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var iptablesPath string = ""
var portMap map[uint16]*CommLogger
var portMutex sync.Mutex
var remoteMap map[string]*CommLogger
var remoteMutex sync.Mutex

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

func GetRemoteLogger(remoteAddr string, port uint16) *CommLogger {
	remoteMutex.Lock()
	if remoteMap == nil {
		remoteMap = make(map[string]*CommLogger)
	}
	key := remoteAddr + "-" + fmt.Sprintf("%d", port)
	l, ok := remoteMap[key]
	if !ok {
		if _, err := os.Stat("./logs"); os.IsNotExist(err) {
			log.Println("Creating logs directory")
			err := os.Mkdir("./logs", 0755)
			if err != nil {
				log.Fatalln(err)
			}
		}
		n, err := NewCommLogger("./logs/"+key+".log", port, remoteAddr)
		if err != nil {
			log.Fatalln(err)
		}
		remoteMap[key] = n
		l = n
	}
	remoteMutex.Unlock()
	return l
}

func getLoggerFromPort(remotePort uint16) *CommLogger {
	portMutex.Lock()
	logger, ok := portMap[uint16(remotePort)]
	if ok != true {
		log.Fatalf("Could not find forward port %d", remotePort)
	}
	portMutex.Unlock()
	return logger
}

func PipeConn(srcConn *net.Conn, destConn *net.Conn, loggingType LogType) {
	var logger *CommLogger = nil
	// We record by remote address
	var remoteAddr string = ""
	remotePort := 0
	// var localAddr string = ""
	var localPort uint16 = 0
	if loggingType == LoggingOutbound {
		// If this the outbound side, the remote IP is in the destination connection
		remoteAddrSplit := strings.Split((*destConn).RemoteAddr().String(), ":")
		remoteAddr = remoteAddrSplit[0]
		rp, err := strconv.Atoi(remoteAddrSplit[1])
		if err != nil {
			remotePort = 0
		}
		remotePort = rp
		// And the local port is on the destination connection too
		localAddrSplit := strings.Split((*destConn).LocalAddr().String(), ":")
		// localAddr = localAddrSplit[0]
		lp, err := strconv.Atoi(localAddrSplit[1])
		if err != nil {
			remotePort = 0
		}
		localPort = uint16(lp)
	} else if loggingType == LoggingInbound {
		// If this is the inbound side, the remote IP is in the source connection
		remoteAddrSplit := strings.Split((*srcConn).RemoteAddr().String(), ":")
		remoteAddr = remoteAddrSplit[0]
		p, err := strconv.Atoi(remoteAddrSplit[1])
		if err != nil {
			remotePort = 0
		}
		remotePort = p
		// And the local port is on the source connection too
		localAddrSplit := strings.Split((*srcConn).LocalAddr().String(), ":")
		// localAddr = localAddrSplit[0]
		lp, err := strconv.Atoi(localAddrSplit[1])
		if err != nil {
			remotePort = 0
		}
		localPort = uint16(lp)
	}

	if remoteAddr != "" {
		if remoteAddr == "127.0.0.1" {
			logger = getLoggerFromPort(uint16(remotePort))
		} else {
			log.Println("Hello")
			logger = GetRemoteLogger(remoteAddr, localPort)
		}
	}

	for {
		(*srcConn).SetReadDeadline(time.Now().Add(20 * time.Second))
		(*destConn).SetWriteDeadline(time.Now().Add(20 * time.Second))
		smallBuffer := []byte{0}
		bytesRead, err := (*srcConn).Read(smallBuffer)
		if err != nil {
			(*srcConn).Close()
			return
		}
		if loggingType == LoggingOutbound {
			logger.WriteOutbound(string(smallBuffer))
		} else if loggingType == LoggingInbound {
			logger.WriteInbound(string(smallBuffer))
		}
		if bytesRead > 0 {
			_, err = (*destConn).Write(smallBuffer)
			if err != nil {
				(*destConn).Close()
				return
			}
		}
	}
}

func AllowTCPPort(port uint16) {
	iptables := getIPtablesPath()
	cmd := exec.Command(iptables, "-I", "OUTPUT", "1", "-w", "-p", "tcp", "--dport", fmt.Sprintf("%d", port), "-j", "ACCEPT")

	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	err := cmd.Run()

	if err != nil {
		log.Println("out:", outBuffer.String(), "err:", errBuffer.String())
		log.Fatal(err)
	}
}

func BlockTCPPort(port uint16) {
	iptables := getIPtablesPath()
	cmd := exec.Command(iptables, "-D", "OUTPUT", "-w", "-p", "tcp", "--dport", fmt.Sprintf("%d", port), "-j", "ACCEPT")

	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer
	err := cmd.Run()

	if err != nil {
		log.Println("out:", outBuffer.String(), "err:", errBuffer.String())
		log.Fatal(err)
	}
}

func AddForwardPort(localPort uint16, destPort uint16, remoteAddr string) {
	portMutex.Lock()
	if portMap == nil {
		portMap = make(map[uint16]*CommLogger)
	}
	portMap[localPort] = GetRemoteLogger(remoteAddr, destPort)
	portMutex.Unlock()
}

func RemoveForwardPort(port uint16, remoteAddr string) {
	portMutex.Lock()
	if portMap == nil {
		portMap = make(map[uint16]*CommLogger)
		return
	}
	delete(portMap, port)
	portMutex.Unlock()
}

func LogInboundData(remoteAddr string, remotePort uint16, port uint16, data string) {
	var logger *CommLogger
	if remoteAddr == "127.0.0.1" {
		logger = getLoggerFromPort(remotePort)
	} else {
		logger = GetRemoteLogger(remoteAddr, port)
	}
	logger.WriteInbound(data)
}
