package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/gob"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

// LoadPublicKey loads an parses a PEM encoded private key file.
func LoadPublicKey(path string) (SignatureValidator, error) {
	gob.Register(rsaPublicKey{})
	dat, err := ioutil.ReadFile(path)
	LogErrorF(err)
	return parsePublicKey(dat)
}

// parsePublicKey parses a PEM encoded private key.
func parsePublicKey(pemBytes []byte) (SignatureValidator, error) {
	gob.Register(rsaPrivateKey{})
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("no key found")
	}

	switch block.Type {
	case "PUBLIC KEY":
		return ParsePublicKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported key block type %q", block.Type)
	}
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
		return nil, errors.New("no key found")
	}

	var signer rsaPrivateKey
	switch block.Type {
	case "RSA PRIVATE KEY":
		rsa, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		signer = rsaPrivateKey{rsa}
	default:
		return nil, fmt.Errorf("unsupported key type %q", block.Type)
	}
	return &signer, nil
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
	Store() ([]byte, error)
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

// Store prepares key for storage
func (r *rsaPublicKey) Store() ([]byte, error) {
	return x509.MarshalPKIXPublicKey(r.PublicKey)
}

// ParsePublicKey read key from raw storage
func ParsePublicKey(data []byte) (SignatureValidator, error) {
	tmpKey, err := x509.ParsePKIXPublicKey(data)
	rsaKey, ok := tmpKey.(*rsa.PublicKey)
	if err != nil || !ok {
		return nil, errors.New("invalid key type, only RSA is supported")
	}
	return &rsaPublicKey{rsaKey}, nil
}
