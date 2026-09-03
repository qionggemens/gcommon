package gutil

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(src []byte) ([]byte, error) {
	length := len(src)
	if length == 0 {
		return nil, errors.New("pkcs5: empty input")
	}
	paddingNum := int(src[length-1])
	if paddingNum == 0 || paddingNum > length {
		return nil, errors.New("pkcs5: invalid padding size")
	}
	for i := 0; i < paddingNum; i++ {
		if src[length-1-i] != byte(paddingNum) {
			return nil, errors.New("pkcs5: invalid padding")
		}
	}
	return src[:length-paddingNum], nil
}

// AesEncryptOfECBWithPKCS5Padding aes 加密（ECB）
func AesEncryptOfECBWithPKCS5Padding(key []byte, origData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	origData = pkcs5Padding(origData, block.BlockSize())
	buffer := bytes.NewBuffer(nil)
	tmpData := make([]byte, block.BlockSize())
	for index := 0; index < len(origData); index += block.BlockSize() {
		block.Encrypt(tmpData, origData[index:index+block.BlockSize()])
		buffer.Write(tmpData)
	}
	return buffer.Bytes(), nil
}

// AesDecryptOfECBWithPKCS5Padding aes 解密（ECB）
func AesDecryptOfECBWithPKCS5Padding(key []byte, origData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(origData) == 0 || len(origData)%block.BlockSize() != 0 {
		return nil, errors.New("aes: ciphertext length invalid")
	}
	buffer := bytes.NewBuffer(nil)
	tmpData := make([]byte, block.BlockSize())
	for index := 0; index < len(origData); index += block.BlockSize() {
		block.Decrypt(tmpData, origData[index:index+block.BlockSize()])
		buffer.Write(tmpData)
	}
	return pkcs5UnPadding(buffer.Bytes())
}

// AesEncryptOfGCMWithNoPadding aes-GCM 加密
func AesEncryptOfGCMWithNoPadding(key []byte, nonce []byte, origData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, err
	}
	return gcm.Seal(nil, nonce, origData, nil), nil
}

// AesDecryptOfGCMWithNoPadding aes-GCM 解密
func AesDecryptOfGCMWithNoPadding(key []byte, nonce []byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}
