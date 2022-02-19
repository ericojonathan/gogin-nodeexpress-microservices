package main

import (
    "fmt"
    "testing"
)

func TestEncryptDecrypt(t *testing.T) {
    
	data := "this is sensitive data"

	encr, err := Encrypt(data)
	if err != nil {
		fmt.Println(err)
	}

	decr, err := Decrypt(encr)
	if err != nil {
		fmt.Println(err)
	}

	if decr != data {
		t.Errorf("Decrypt = %s; want %s", decr, data)
	}	
}

