package services

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func (s *HTTPService) handleRequest(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Session not available\n")
}

type HTTPService struct {
	port uint16
}

func (s *HTTPService) Start(address string, port uint16) error {
	log.Printf("Starting HTTP server at %d\n", port)
	s.port = port
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
