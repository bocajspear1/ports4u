package services

type Service interface {
	Start(address string, port uint16) error
}
