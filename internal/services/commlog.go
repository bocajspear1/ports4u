package services

import (
	"log"
	"os"
	"strconv"
	"sync"
)

type LogType int16

const (
	LoggingOutbound LogType = iota
	LoggingInbound
	LoggingNone
)

type CommLogger struct {
	logOutbound bool
	logMutex    sync.Mutex
	fileHandle  *os.File
	remoteAddr  string
}

func escapeString(data string) string {
	newString := strconv.QuoteToASCII(data)
	return newString[1:len(newString)-1] + "\n"
}

func (c *CommLogger) WriteOutbound(data string) {
	c.logMutex.Lock()
	if !c.logOutbound {
		c.fileHandle.WriteString(">>>>>>>> " + c.remoteAddr + " ----------------------------\n")
		c.fileHandle.Sync()
		c.logOutbound = true
	}
	c.fileHandle.WriteString(escapeString(data))
	c.fileHandle.Sync()

	c.logMutex.Unlock()
}

func (c *CommLogger) WriteInbound(data string) {
	c.logMutex.Lock()
	if c.logOutbound {
		c.fileHandle.WriteString("<<<<<<<< " + c.remoteAddr + " ----------------------------\n")
		c.fileHandle.Sync()
		c.logOutbound = false
	}
	c.fileHandle.WriteString(escapeString(data))
	c.fileHandle.Sync()

	c.logMutex.Unlock()
}

func (c *CommLogger) GetRemoteAddr() string {
	return c.remoteAddr
}

func NewCommLogger(path string, port uint16, remoteAddr string) (*CommLogger, error) {
	log.Printf("New logger for %s:%d", remoteAddr, port)
	cl := new(CommLogger)

	handle, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	cl.fileHandle = handle
	cl.remoteAddr = remoteAddr
	cl.logOutbound = true

	return cl, nil
}
