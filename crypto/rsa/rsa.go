package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
)

func genKeyPair() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		fmt.Println(err)
		return
	}

	privateFile(privateKey)
	publicFile(&privateKey.PublicKey)
}

func privateFile(privateKey *rsa.PrivateKey) {
	fmt.Println("privateKey len: ", privateKey.N.BitLen())
	derPri := x509.MarshalPKCS1PrivateKey(privateKey)
	pemPri := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derPri,
	}

	filePri, err := os.Create("private.pem")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer filePri.Close()

	if err = pem.Encode(filePri, pemPri); err != nil {
		fmt.Println(err)
		return
	}
}

func publicFile(publicKey *rsa.PublicKey) {
	fmt.Println("publicKey len:", publicKey.N.BitLen())
	derPub, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	pemPub := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: derPub,
	}

	filePub, err := os.Create("public.pem")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer filePub.Close()

	if err = pem.Encode(filePub, pemPub); err != nil {
		fmt.Println(err)
		return
	}
}

func rsaEncrypt(msg []byte) ([]byte, error) {
	data, err := ioutil.ReadFile("public.pem")
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("public.pem bad")
	}

	pubItf, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pub := pubItf.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, msg)
}

func rsaDecrypt(ciphertext []byte) ([]byte, error) {
	data, err := ioutil.ReadFile("private.pem")
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("private.pem bad")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
}

func encryptDecrypt() {
	ciphertext, err := rsaEncrypt([]byte(`1234`))
	if err != nil {
		fmt.Println(err)
		return
	}

	msg, err := rsaDecrypt(ciphertext)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(msg))
}

func sign(msg []byte) ([]byte, error) {
	data, err := ioutil.ReadFile("private.pem")
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("private.pem bad")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	h.Write(msg)

	return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, h.Sum(nil))
}

func verify(msg, signature []byte) error {
	data, err := ioutil.ReadFile("public.pem")
	if err != nil {
		return err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return fmt.Errorf("public.pem bad")
	}

	pubItf, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	pub := pubItf.(*rsa.PublicKey)
	//pub.N.BitLen()

	h := sha256.New()
	h.Write(msg)

	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, h.Sum(nil), signature)
}

func signVerify() {
	msg := []byte("zz hello")
	signature, err := sign(msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(string(signature))
	fmt.Println(verify(msg, signature))
}

//func main() {
//	genKeyPair()
//	encryptDecrypt()
//	signVerify()
//}
