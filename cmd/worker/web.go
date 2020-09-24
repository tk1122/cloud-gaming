package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"html/template"
	"math/big"
	"net/http"
	"os"
	"time"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func genPem() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	must(err)

	SNLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	SN, err := rand.Int(rand.Reader, SNLimit)
	must(err)

	certificate := x509.Certificate{
		SerialNumber: SN,
		Subject: pkix.Name{
			Organization: []string{"tk1122"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	certificate.DNSNames = append(certificate.DNSNames, "localhost")
	certificate.EmailAddresses = append(certificate.EmailAddresses, "kaka.ngo@gmail.com")

	certBytes, err := x509.CreateCertificate(rand.Reader, &certificate, &certificate, &privateKey.PublicKey, privateKey)
	must(err)

	certFile, err := os.Create("cert.pem")
	must(err)
	defer func() {
		must(certFile.Close())
	}()
	must(pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}))

	keyFile, err := os.OpenFile("key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	must(err)
	defer func() {
		must(keyFile.Close())
	}()
	must(pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}))
}

func getWeb(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./web/index.html")
	must(err)
	must(t.Execute(w, nil))
}
