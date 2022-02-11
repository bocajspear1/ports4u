package identify

import "bytes"

func IsHTTP(data []byte) bool {
	return bytes.Contains(data, []byte("HTTP/"))
}

func IsTLS(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0x16, 0x03, 0x01})
}
