package services

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
)

func (s *HTTPService) handleRequest(w http.ResponseWriter, req *http.Request) {

	remoteAddrSplit := strings.Split(req.RemoteAddr, ":")
	remoteAddr := remoteAddrSplit[0]
	if remoteAddr != "127.0.0.1" {
		inBytes, err := httputil.DumpRequest(req, true)
		if err == nil {
			logger := GetRemoteLogger(remoteAddr, s.port)
			logger.WriteInbound(string(inBytes))
		}

	}

	fmt.Fprintf(w, "<html><head><title>Session Invalid></title></head><body>Session not available</body></html>\n")
}

type HTTPService struct {
	port uint16
}

func (s *HTTPService) Start(address string, port uint16) error {
	log.Printf("Starting HTTP server at %d\n", port)
	s.port = port
	AllowTCPPort(s.port)
	http.HandleFunc("/", s.handleRequest)

	err := http.ListenAndServe(":"+strconv.Itoa(int(s.port)), nil)
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func NewHTTPService() *HTTPService {
	service := new(HTTPService)
	return service
}
