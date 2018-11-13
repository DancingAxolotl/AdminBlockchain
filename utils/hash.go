package utils

import (
	"bytes"
	"crypto/sha256"
)

// Hash produces a sha256 hash of a list of parameters
func Hash(params ...interface{}) []byte {
	hash := sha256.New()
	var buffer bytes.Buffer
	for i, item := range params {
		s, ok := item.(string)
		if ok {
			buffer.WriteString(s)
		} else {
			buffer.WriteString(string(i * 17))
		}
	}
	hash.Write(buffer.Bytes())
	return hash.Sum(nil)
}
