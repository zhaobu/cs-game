package aescrc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func AesCbcEncrypto(plaintext, key []byte, iv []byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext = pkcs5Padding(plaintext, aes.BlockSize)

	if iv == nil {
		iv = make([]byte, aes.BlockSize)
		_, err = io.ReadFull(rand.Reader, iv)
		if err != nil {
			return nil, err
		}
	}

	blockMode := cipher.NewCBCEncrypter(block, iv)

	// ciphertext = make([]byte, len(iv)+len(plaintext))
	// copy(ciphertext, iv)
	// blockMode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	ciphertext = make([]byte, len(plaintext))
	blockMode.CryptBlocks(ciphertext, plaintext)
	return
}

func AesCbcDecrypto(ciphertext, key []byte, iv []byte) (plaintext []byte, err error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("Cipher data length less than aes block size")
	}

	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("bad blocksize(%v), aes.BlockSize = %v\n", len(ciphertext), aes.BlockSize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if iv == nil {
		iv = ciphertext[:aes.BlockSize]
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)

	// plaintext = make([]byte, len(ciphertext)-aes.BlockSize)
	// blockMode.CryptBlocks(plaintext, ciphertext[aes.BlockSize:])
	plaintext = make([]byte, len(ciphertext))
	blockMode.CryptBlocks(plaintext, ciphertext)
	return pkcs5UnPadding(plaintext), nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
