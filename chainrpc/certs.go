package chainrpc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path"
	"time"
)

var (
	CA      = "ca.pem"
	CaKey   = "ca_key.pem"
	Cert    = "cert.pem"
	CertKey = "cert_key.pem"
)

var baseCA = &x509.Certificate{
	SerialNumber: big.NewInt(2020),
	Subject: pkix.Name{
		Organization:  []string{"Ogen cert"},
		Country:       []string{"Crypto"},
		Province:      []string{""},
		Locality:      []string{""},
		StreetAddress: []string{""},
		PostalCode:    []string{""},
	},
	NotBefore:             time.Now(),
	NotAfter:              time.Now().AddDate(10, 0, 0),
	IsCA:                  true,
	ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	BasicConstraintsValid: true,
}

// LoadCerts will attempt to load certificates previously generatet. If it fails it will generate them and load them again.
func LoadCerts(dataFolder string) (*x509.CertPool, error) {
open:
	ca, err := ioutil.ReadFile(path.Join(dataFolder, "cert", CA))
	if err != nil {
		err := GenerateCerts(dataFolder)
		if err != nil {
			return nil, err
		}
		goto open
	}
	caKey, err := ioutil.ReadFile(path.Join(dataFolder, "cert", CaKey))
	if err != nil {
		err := GenerateCerts(dataFolder)
		if err != nil {
			return nil, err
		}
		goto open
	}
	cert, err := ioutil.ReadFile(path.Join(dataFolder, "cert", Cert))
	if err != nil {
		err := GenerateCerts(dataFolder)
		if err != nil {
			return nil, err
		}
		goto open
	}
	certKey, err := ioutil.ReadFile(path.Join(dataFolder, "cert", CertKey))
	if err != nil {
		err := GenerateCerts(dataFolder)
		if err != nil {
			return nil, err
		}
		goto open
	}
	caBlock, _ := pem.Decode(ca)
	caKeyBlock, _ := pem.Decode(caKey)
	certBlock, _ := pem.Decode(cert)
	certKeyBlock, _ := pem.Decode(certKey)
	if caBlock.Type != "CERTIFICATE" ||
		certBlock.Type != "CERTIFICATE" ||
		caKeyBlock.Type != "RSA PRIVATE KEY" ||
		certKeyBlock.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("wrong file format loaded")
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, errors.New("failed to append certificate to certpool")
	}
	return certPool, nil
}

// GenerateCerts will generate a CA and a certificate for the gRPC tls transport and https.
func GenerateCerts(dataFolder string) error {
	os.Mkdir(path.Join(dataFolder, "cert"), 0777)
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, baseCA, baseCA, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}
	caFile, err := os.Create(path.Join(dataFolder, "cert", CA))
	if err != nil {
		return err
	}
	pem.Encode(caFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	caPriv, err := os.Create(path.Join(dataFolder, "cert", CaKey))
	if err != nil {
		return err
	}
	pem.Encode(caPriv, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	certData := baseCA
	certData.IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	certData.SubjectKeyId = []byte{1, 2, 3, 4, 6}
	certData.IsCA = false
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, certData, baseCA, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	certFile, err := os.Create(path.Join(dataFolder, "cert", Cert))
	if err != nil {
		return err
	}
	pem.Encode(certFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certKey, err := os.Create(path.Join(dataFolder, "cert", CertKey))
	if err != nil {
		return err
	}
	pem.Encode(certKey, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	return nil
}
