package services

import (
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bocajspear1/ports4u/internal/identify"
)

type PortService struct {
	port      uint16
	connMutex sync.Mutex
	connCount uint64
	listener  net.Listener
	active    bool
}

func (s *PortService) forwardData(startData []byte, origPort uint16, destPort uint16, lastConn *net.Conn, cleartext bool) {
	log.Printf("Forwarding to port %d\n", destPort)
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(int(destPort)))
	if err != nil {
		log.Println(err)
		return
	}

	remoteAddrSplit := strings.Split((*lastConn).RemoteAddr().String(), ":")
	remoteAddr := remoteAddrSplit[0]

	localAddrSplit := strings.Split(conn.LocalAddr().String(), ":")
	localPort, err := strconv.Atoi(localAddrSplit[1])
	if err != nil {
		log.Println(err)
		return
	}

	// Add the forward source port to a map so any service down the chain
	// can convert to remote address
	AddForwardPort(uint16(localPort), origPort, remoteAddr)

	conn.Write(startData)

	// Only log data is we know we are cleartext
	if cleartext {
		go PipeConn(&conn, lastConn, LoggingOutbound)
		PipeConn(lastConn, &conn, LoggingInbound)
	} else {
		go PipeConn(&conn, lastConn, LoggingNone)
		PipeConn(lastConn, &conn, LoggingNone)
	}

	log.Printf("Finished forwarded connection\n")
	RemoveForwardPort(uint16(localPort), remoteAddr)

	s.connMutex.Lock()
	s.connCount -= 1
	s.connMutex.Unlock()
}

func (s *PortService) IsActive() bool {
	return s.active
}

func (s *PortService) tcpHandler(port uint16, conn net.Conn) {

	addrSplit := strings.Split(conn.RemoteAddr().String(), ":")

	remoteAddr := addrSplit[0]
	remotePort, err := (strconv.Atoi(addrSplit[1]))
	if err != nil {
		remotePort = 0
	}

	const chunkSize = 512

	connDied := false
	forwarded := false
	finalBuffer := make([]byte, 0)

	// var outFile *os.File
	// var outPath string

	for !connDied && !forwarded {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		smallBuffer := make([]byte, chunkSize)

		bytesRead, err := conn.Read(smallBuffer)

		if bytesRead > 0 {
			log.Printf("Read %d bytes\n", bytesRead)
			finalBuffer = append(finalBuffer, smallBuffer[0:bytesRead]...)

			if identify.IsHTTP(finalBuffer) {
				log.Println("Think I found HTTP traffic!")
				// Be sure to log the chunk of data we already got
				LogInboundData(remoteAddr, uint16(remotePort), port, string(smallBuffer[0:bytesRead]))
				go s.forwardData(finalBuffer, port, 80, &conn, true)
				forwarded = true
			} else if identify.IsTLS(finalBuffer) {
				log.Println("Think I found SSL/TLS traffic!")
				go s.forwardData(finalBuffer, port, 443, &conn, false)
				forwarded = true
			} else {
				LogInboundData(remoteAddr, uint16(remotePort), port, string(smallBuffer[0:bytesRead]))
			}

		}

		if err != nil {
			// Maybe do something if the error is of a certian type?
			connDied = true
		}
	}

	if !forwarded {
		conn.Close()

		log.Printf("Finished connection from %s:%d", remoteAddr, remotePort)
		s.connMutex.Lock()
		s.connCount -= 1
		s.connMutex.Unlock()
	} else {
		log.Printf("Forwarding connection from %s:%d", remoteAddr, remotePort)
	}

}

func (s *PortService) timeoutAccept() {
	time.Sleep(30 * time.Second)
	for {
		s.connMutex.Lock()
		if s.connCount == 0 {
			log.Printf("Closing port service %d\n", s.port)
			s.active = false
			s.listener.Close()
			BlockTCPPort(s.port)
			return
		}
		s.connMutex.Unlock()
		time.Sleep(30 * time.Second)
	}
}

func (s *PortService) Start(address string, port uint16) error {
	s.active = true
	s.port = port
	var err error
	s.listener, err = net.Listen("tcp", ":"+strconv.Itoa(int(s.port)))
	if err != nil {
		log.Fatalf("Failed to start port service on port %d: %s\n", s.port, err)
		return nil
	}

	AllowTCPPort(s.port)
	defer s.listener.Close()

	go s.timeoutAccept()

	for {
		conn, aerr := s.listener.Accept()
		s.connMutex.Lock()
		s.connCount += 1
		s.connMutex.Unlock()

		if aerr != nil {
			log.Printf("Failed connection")
			return nil
		}

		go s.tcpHandler(s.port, conn)
	}
}

func NewPortService() *PortService {
	service := new(PortService)
	service.active = false
	return service
}
