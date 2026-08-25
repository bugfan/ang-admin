package service

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

// GenerateSelfSignedCert creates a leaf RSA certificate signed by a generated Root CA, returning private key, cert and CA/intermediate cert
func GenerateSelfSignedCert(commonName string, dnsNames []string, validDays int) (string, string, string, error) {
	if commonName == "" {
		commonName = "local.i443.cn"
	}
	if validDays <= 0 {
		validDays = 365
	}

	notBefore := time.Now().Add(-10 * time.Minute)
	notAfter := notBefore.Add(time.Duration(validDays) * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	// 1. 生成 Root CA 证书与私钥 (作为中间/根证书供客户端信任)
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", err
	}

	caSerial, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", "", err
	}

	caTemplate := x509.Certificate{
		SerialNumber: caSerial,
		Subject: pkix.Name{
			Organization: []string{"ANG Root Authority"},
			CommonName:   "ANG Root CA",
		},
		NotBefore:             notBefore,
		NotAfter:              notBefore.Add(time.Duration(validDays*2) * 24 * time.Hour), // CA 有效期更长
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caDerBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return "", "", "", err
	}

	caCertBuf := new(bytes.Buffer)
	_ = pem.Encode(caCertBuf, &pem.Block{Type: "CERTIFICATE", Bytes: caDerBytes})
	caPEM := caCertBuf.String()

	// 2. 生成服务器站点私钥 (Server Private Key)
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", err
	}

	serverSerial, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", "", err
	}

	serverTemplate := x509.Certificate{
		SerialNumber: serverSerial,
		Subject: pkix.Name{
			Organization: []string{"ANG Server Certificate"},
			CommonName:   commonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
	}

	serverDerBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, &caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		return "", "", "", err
	}

	serverCertBuf := new(bytes.Buffer)
	_ = pem.Encode(serverCertBuf, &pem.Block{Type: "CERTIFICATE", Bytes: serverDerBytes})
	serverCertPEM := serverCertBuf.String()

	serverKeyBuf := new(bytes.Buffer)
	_ = pem.Encode(serverKeyBuf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})
	serverKeyPEM := serverKeyBuf.String()

	return serverKeyPEM, serverCertPEM, caPEM, nil
}
