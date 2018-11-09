package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

// LoadPublicKey loads an parses a PEM encoded private key file.
func LoadPublicKey(path string) (SignatureValidator, error) {
	dat, err := ioutil.ReadFile(path)
	LogErrorF(err)

	return parsePublicKey(dat)
}

// parsePublicKey parses a PEM encoded private key.
func parsePublicKey(pemBytes []byte) (SignatureValidator, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("ssh: no key found")
	}

	var rawkey interface{}
	switch block.Type {
	case "PUBLIC KEY":
		rsa, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rawkey = rsa
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %q", block.Type)
	}

	return newValidatorFromKey(rawkey)
}

// LoadPrivateKey loads an parses a PEM encoded private key file.
func LoadPrivateKey(path string) (SignatureCreator, error) {
	dat, err := ioutil.ReadFile(path)
	LogErrorF(err)

	return parsePrivateKey(dat)
}

// parsePublicKey parses a PEM encoded private key.
func parsePrivateKey(pemBytes []byte) (SignatureCreator, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("ssh: no key found")
	}

	var rawkey interface{}
	switch block.Type {
	case "RSA PRIVATE KEY":
		rsa, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rawkey = rsa
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %q", block.Type)
	}
	return newSignerFromKey(rawkey)
}

// SignatureCreator creates signatures from a private key.
type SignatureCreator interface {
	// Sign returns raw signature for data.
	Sign(data []byte) ([]byte, error)
}

// SignatureValidator verifies signatures using a public key.
type SignatureValidator interface {
	// CheckSignature checks the signature for data.
	CheckSignature(data []byte, sig []byte) error
}

func newSignerFromKey(k interface{}) (SignatureCreator, error) {
	var sshKey SignatureCreator
	switch t := k.(type) {
	case *rsa.PrivateKey:
		sshKey = &rsaPrivateKey{t}
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %T", k)
	}
	return sshKey, nil
}

func newValidatorFromKey(k interface{}) (SignatureValidator, error) {
	var sshKey SignatureValidator
	switch t := k.(type) {
	case *rsa.PublicKey:
		sshKey = &rsaPublicKey{t}
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %T", k)
	}
	return sshKey, nil
}

type rsaPublicKey struct {
	*rsa.PublicKey
}

type rsaPrivateKey struct {
	*rsa.PrivateKey
}

// Sign signs data
func (r *rsaPrivateKey) Sign(data []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(data)
	d := h.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, r.PrivateKey, crypto.SHA256, d)
}

// CheckSignature verifies the message using the signature
func (r *rsaPublicKey) CheckSignature(message []byte, sig []byte) error {
	h := sha256.New()
	h.Write(message)
	d := h.Sum(nil)
	return rsa.VerifyPKCS1v15(r.PublicKey, crypto.SHA256, d, sig)
}
