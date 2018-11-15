package utils

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// Hash produces a sha256 hash of a list of parameters
func Hash(params ...interface{}) []byte {
	hash := sha256.New()
	var buffer bytes.Buffer
	for _, item := range params {
		buffer.WriteString(fmt.Sprintf("%v", item))
	}
	hash.Write(buffer.Bytes())
	return hash.Sum(nil)
}
