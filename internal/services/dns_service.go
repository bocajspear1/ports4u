package services

import (
	"log"
	"net"
	"os"
	"strconv"

	"github.com/miekg/dns"
)

func (s *DNSService) handleRequest(wr dns.ResponseWriter, query *dns.Msg) {
	message := new(dns.Msg)
	message.SetReply(query)
	message.Compress = false
	message.RecursionAvailable = true

	switch query.Opcode {
	case dns.OpcodeQuery:
		for _, question := range query.Question {
			switch question.Qtype {
			case dns.TypeA:
				log.Printf("Query for %s\n", question.Name)

				CheckLogDir()

				// Log all DNS requests
				out, err := os.OpenFile("./logs/domains.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Fatal("Failed to open ./logs/domains.txt")
				}
				_, err = out.Write([]byte(question.Name + "\n"))
				if err != nil {
					log.Fatal("Failed to write to ./logs/domains.txt")
				}
				out.Close()

				// Make response
				aRec := &dns.A{
					Hdr: dns.RR_Header{
						Name:   question.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    3600,
					},
					A: net.ParseIP(s.defaultResponse),
				}
				message.Answer = append(message.Answer, aRec)
			}
		}
	}

	wr.WriteMsg(message)
}

type DNSService struct {
	port            uint16
	defaultResponse string
}

func (s *DNSService) Start(address string, port uint16) error {

	dns.HandleFunc(".", s.handleRequest)

	s.defaultResponse = address

	server := &dns.Server{Addr: ":" + strconv.Itoa(int(port)), Net: "udp"}
	log.Printf("Starting DNS server at %d\n", port)
	log.Printf("All responses are %s\n", s.defaultResponse)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}

	return nil
}

func NewDNSService() *DNSService {
	service := new(DNSService)
	return service
}
