package services

import (
	"crypto/tls"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/bocajspear1/ports4u/internal/identify"
)

type TLSService struct {
	port uint16
}

func (s *TLSService) forwardData(startData []byte, destPort uint16, lastConn *net.Conn) {
	log.Printf("TLS forwarding to port %d\n", destPort)
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(int(destPort)))
	if err != nil {
		log.Println(err)
		return
	}

	conn.Write(startData)
	go PipeConn(&conn, lastConn, LoggingOutbound)
	PipeConn(lastConn, &conn, LoggingInbound)

	log.Printf("Finished forwarded TLS connection\n")
}

func (s *TLSService) tlsTCPHandler(port uint16, conn net.Conn) {

	addrSplit := strings.Split(conn.RemoteAddr().String(), ":")

	remoteAddr := addrSplit[0]
	remotePort, err := (strconv.Atoi(addrSplit[1]))
	if err != nil {
		remotePort = 0
	}

	log.Printf("Got TLS connection from %s:%d", remoteAddr, remotePort)

	const chunkSize = 512

	connDied := false
	forwarded := false
	finalBuffer := make([]byte, 0)

	for !connDied && !forwarded {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		smallBuffer := make([]byte, chunkSize)

		bytesRead, err := conn.Read(smallBuffer)

		if bytesRead > 0 {
			log.Printf("Read %d bytes\n", bytesRead)
			finalBuffer = append(finalBuffer, smallBuffer[0:bytesRead]...)

			if identify.IsHTTP(finalBuffer) {
				log.Println("Think I found HTTP traffic!")
				LogInboundData(remoteAddr, uint16(remotePort), port, string(smallBuffer[0:bytesRead]))
				go s.forwardData(finalBuffer, 80, &conn)
				forwarded = true
			} else {
				LogInboundData(remoteAddr, uint16(remotePort), port, string(smallBuffer[0:bytesRead]))
			}
		}

		if err != nil {
			log.Println(err)
			// Maybe do something if the error is of a certian type?
			connDied = true
		}
	}

	if !forwarded {
		conn.Close()
		log.Printf("Finished TLS connection from %s:%d", remoteAddr, remotePort)
	} else {
		log.Printf("Forwarding TLS connection from %s:%d", remoteAddr, remotePort)
	}

}

func (s *TLSService) Start(address string, port uint16) error {
	log.Printf("Starting TLS server at %d\n", port)
	s.port = port

	certs, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Println(err)
		return err
	}
	config := &tls.Config{Certificates: []tls.Certificate{certs}}
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(int(s.port)), config)
	if err != nil {
		log.Fatalf("Failed to start port service on port %d: %s\n", s.port, err)
		return nil
	}

	AllowTCPPort(s.port)
	defer listener.Close()

	for {
		conn, aerr := listener.Accept()

		if aerr != nil {
			log.Printf("Failed connection")
			return nil
		}

		go s.tlsTCPHandler(s.port, conn)
	}
}

func NewTLSService() *TLSService {
	service := new(TLSService)
	return service
}
