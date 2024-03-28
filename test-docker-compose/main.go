package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if err := makeCerts(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func makeCerts() error {
	now := time.Now()

	serialNumber := big.NewInt(1024)
	ca := &x509.Certificate{
		SerialNumber:          serialNumber,
		BasicConstraintsValid: true,
		Subject: pkix.Name{
			Organization: []string{"Raintank, Inc."},
		},
		DNSNames: []string{
			"grafana",
			"mtls-proxy",
		},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		NotBefore:   now,
		NotAfter:    now.Add(1 * time.Hour),
		IsCA:        true,
	}

	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	if err := writePEMFiles("ca", caBytes, caPrivKey); err != nil {
		return err
	}

	// Create client and server certificates
	for _, name := range []string{"client", "grafana"} {
		serialNumber = serialNumber.Add(serialNumber, big.NewInt(1))
		// copy CA data
		crt := &x509.Certificate{}
		*crt = *ca

		// overwrite CA data that's not needed for certificates
		crt.Subject.CommonName = name
		crt.SerialNumber = serialNumber
		crt.SubjectKeyId = []byte{1, 2, 3, 4, 6}
		crt.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
		crt.IsCA = false
		crt.IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv4(0, 0, 0, 0), net.IPv6loopback}

		crtPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return fmt.Errorf("cannot generate RSA key for certificate %s: %w", name, err)
		}

		crtBytes, err := x509.CreateCertificate(rand.Reader, crt, ca, &crtPrivKey.PublicKey, caPrivKey)
		if err != nil {
			return fmt.Errorf("cannot create certificate %s: %w", name, err)
		}

		if err := writePEMFiles(name, crtBytes, crtPrivKey); err != nil {
			return err
		}

	}

	return nil
}

func writePEMFiles(name string, crtBytes []byte, crtPrivKey *rsa.PrivateKey) error {
	os.MkdirAll("certs", 0755)
	if err := writePEMFile(fmt.Sprintf("certs/%s.crt", name), "CERTIFICATE", crtBytes); err != nil {
		return fmt.Errorf("cannot write certificate file: %w", err)
	}

	if err := writePEMFile(fmt.Sprintf("certs/%s.key", name), "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(crtPrivKey)); err != nil {
		return fmt.Errorf("cannot write key file: %w", err)
	}

	return nil
}

func writePEMFile(name, pemType string, data []byte) error {
	buf := new(bytes.Buffer)

	err := pem.Encode(buf, &pem.Block{
		Type:  pemType,
		Bytes: data,
	})
	if err != nil {
		return fmt.Errorf("cannot PEM encode %s: %w", pemType, err)
	}

	name, err = filepath.Abs(name)
	if err != nil {
		return err
	}

	err = os.WriteFile(name, buf.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("cannot write to PEM file %s: %w", name, err)
	}

	return nil
}
