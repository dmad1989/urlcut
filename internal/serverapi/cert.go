package serverapi

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
	"time"
)

func CreateCert(certPath, keyPath string) error {
	// создаём шаблон сертификата
	cert := &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(1658),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"Yandex.Praktikum"},
			Country:      []string{"RU"},
		},
		// разрешаем использование сертификата для 127.0.0.1 и ::1
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата — 10 лет
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	// создаём новый приватный RSA-ключ длиной 4096 бит
	// обратите внимание, что для генерации ключа и сертификата
	// используется rand.Reader в качестве источника случайных данных
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("CreateCert: GenerateKey: %w", err)
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("CreateCert: CreateCertificate: %w", err)
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	err = saveToFile(certPath, "CERTIFICATE", certBytes)
	if err != nil {
		return fmt.Errorf("CreateCert: CERTIFICATE: %w", err)
	}

	err = saveToFile(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(privateKey))
	if err != nil {
		return fmt.Errorf("CreateCert: RSA PRIVATE KEY: %w", err)
	}

	return nil
}

func saveToFile(path string, t string, b []byte) error {
	var (
		buf bytes.Buffer
		f   *os.File
	)
	err := pem.Encode(&buf, &pem.Block{
		Type:  t,
		Bytes: b,
	})
	if err != nil {
		return fmt.Errorf("saveToFile: pem.Encode: %w", err)
	}

	f, err = os.Create(path)
	defer func() (err error) {
		err = f.Close()
		if err != nil {
			err = fmt.Errorf("saveToFile: file close: %w", err)
		}
		return
	}()

	if err != nil {
		return fmt.Errorf("saveToFile: path Create: %w", err)
	}
	_, err = buf.WriteTo(f)
	if err != nil {
		return fmt.Errorf("saveToFile: buf.WriteTo: %w", err)
	}

	return nil
}
