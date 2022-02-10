package services

import (
	"bytes"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PortService struct {
	port      uint16
	connMutex sync.Mutex
	connCount uint64
	listener  net.Listener
	active    bool
}

func (s *PortService) forwardData(startData []byte, destPort uint16, lastConn *net.Conn) {
	log.Printf("Forwarding to port %d\n", destPort)
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(int(destPort)))
	if err != nil {
		log.Println(err)
		return
	}
	conn.Write(startData)
	go PipeConn(&conn, lastConn)
	PipeConn(lastConn, &conn)
	log.Printf("Finished forwarded connection\n")
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
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	for !connDied && !forwarded {
		smallBuffer := make([]byte, chunkSize)

		bytesRead, err := conn.Read(smallBuffer)

		if bytesRead > 0 {
			log.Printf("Read %d bytes\n", bytesRead)
			finalBuffer = append(finalBuffer, smallBuffer[0:bytesRead]...)

			if bytes.Contains(finalBuffer, []byte("HTTP/")) {
				log.Println("Think I found HTTP traffic!")
				go s.forwardData(finalBuffer, 80, &conn)
				forwarded = true
			} else if bytes.HasPrefix(finalBuffer, []byte{0x16, 0x03, 0x01}) {
				log.Println("Think I found SSL/TLS traffic!")
				go s.forwardData(finalBuffer, 443, &conn)
				forwarded = true
			}
			// bytesReadTotal += bytesRead
			// if bytesReadTotal < toFileSize {
			// 	finalBuffer = append(finalBuffer, smallBuffer[0:bytesRead]...)
			// } else {
			// 	if outFile == nil {
			// 		outPath = "./large/tcp-" + strconv.Itoa(port) + "-" + strconv.FormatInt(time.Now().Unix(), 10) + ".large"
			// 		outFile, err = os.OpenFile(outPath, os.O_RDWR|os.O_CREATE, 0444)
			// 		if err != nil {
			// 			log.Printf("Could not open large file: %s\n", err)
			// 			conn.Close()
			// 			connDied = true
			// 		}
			// 		defer outFile.Close()
			// 		outFile.Write(finalBuffer)
			// 	}

			// 	outFile.Write(smallBuffer[0:bytesRead])
			// }

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

	return nil
}

func NewPortService() *PortService {
	service := new(PortService)
	service.active = false
	return service
}

func SpawnPortService(address string, port uint16) {
	portSvc := NewPortService()
	portSvc.Start(address, port)
}
