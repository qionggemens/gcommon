package gutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/qionggemens/gcommon/pkg/glog"
)

// RSAEncrypt RSA 加密；失败返回 error（不再 panic）
func RSAEncrypt(plainTextBytes []byte, publicKeyBytes []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKeyBytes)
	if block == nil {
		return nil, errors.New("rsa encrypt: invalid public key pem")
	}
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("rsa encrypt: not rsa public key")
	}
	return rsa.EncryptPKCS1v15(rand.Reader, publicKey, plainTextBytes)
}

// RSADecrypt RSA 解密；失败返回 error（不再 panic / 忽略错误）
func RSADecrypt(cipherTextBytes []byte, privateKeyBytes []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return nil, errors.New("rsa decrypt: invalid private key pem")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherTextBytes)
}

// RSAGenerate 密钥生成
func RSAGenerate(bits int) ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		glog.Errorf("RSAGenerate fail - msg:%s", err.Error())
		return nil, nil, errors.New("rsa generate fail")
	}
	X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)
	privateBlock := pem.Block{Type: "RSA Private Key", Bytes: X509PrivateKey}
	prk := pem.EncodeToMemory(&privateBlock)
	if prk == nil {
		glog.Errorf("RSAGenerate fail - msg:generate private key fail")
		return nil, nil, errors.New("generate private key fail")
	}
	publicKey := privateKey.PublicKey
	X509PublicKey, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		glog.Errorf("RSAGenerate fail - msg:%s", err.Error())
		return nil, nil, errors.New("rsa generate fail")
	}
	publicBlock := pem.Block{Type: "RSA Public Key", Bytes: X509PublicKey}
	puk := pem.EncodeToMemory(&publicBlock)
	if puk == nil {
		glog.Errorf("RSAGenerate fail - msg:generate public key fail")
		return nil, nil, errors.New("generate public key fail")
	}
	return prk, puk, nil
}
