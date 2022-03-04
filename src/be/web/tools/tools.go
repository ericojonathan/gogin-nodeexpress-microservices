package web

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"			

	"github.com/mergermarket/go-pkcs7"
	"github.com/spf13/viper"
	
)

var (	
	CIPHER_KEY string
	Conf *viper.Viper
)

func init() {
	
	//ROOT_DIR := "/home/eric/job_tests/k.digital_gogin_nodeexpress/src/be"
	Conf = viper.New()
	Conf.SetConfigFile("../config/default.json")		
	//Conf.SetConfigFile(ROOT_DIR + "/config/default.json")	

	err := Conf.ReadInConfig()	
	if err != nil {
		fmt.Println(err)
	}

	CIPHER_KEY = Conf.Get("crypto.cipher_key").(string)
	if CIPHER_KEY == "" {
		fmt.Println("Cannot get cipher key!")
	}
}

func Encrypt(unencrypted string) (string, error) {
	key := []byte(CIPHER_KEY)
	plainText := []byte(unencrypted)
	plainText, err := pkcs7.Pad(plainText, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf(`plainText: "%s" has error`, plainText)
	}
	if len(plainText)%aes.BlockSize != 0 {
		err := fmt.Errorf(`plainText: "%s" has the wrong block size`, plainText)
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainText)

	return fmt.Sprintf("%x", cipherText), nil
}

func Decrypt(encrypted string) (string, error) {
	key := []byte(CIPHER_KEY)
	cipherText, _ := hex.DecodeString(encrypted)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	if len(cipherText) < aes.BlockSize {
		panic("cipherText too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]
	if len(cipherText)%aes.BlockSize != 0 {
		panic("cipherText is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)

	cipherText, _ = pkcs7.Unpad(cipherText, aes.BlockSize)
	return fmt.Sprintf("%s", cipherText), nil
}