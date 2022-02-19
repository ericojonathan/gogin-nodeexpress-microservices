package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"	
	"github.com/mergermarket/go-pkcs7"	
	"github.com/spf13/viper"
	"io"
	"log"
	
    _ "github.com/go-sql-driver/mysql"
	//"log"
	// "reflect"
)

var (
	CIPHER_KEY string
	Conf *viper.Viper
)

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

func test_encr_msg(c *gin.Context) {

	var data = "this is sensitive data"
	data_encr, _ := Encrypt(data)
	
	c.JSON(200, gin.H{
		"message": data_encr,
	})
}

type Employee struct {
	ID int `json:"id"`
	Job_Title string `json:"job_title"`
	Email_Address string `json:"email_address"`
	FirstName_LastName string `json:"firstName_LastName"`
}

var (
	ctx context.Context
	db  *sql.DB
)

//TODO: Get single function that returns db
//func get_db() (db.SQLDB)

func get_users_encr(c *gin.Context) {
	
	fmt.Println("[GET USERS ENCRYPT]")

	db_user := Conf.Get("mysql.user")
	db_pass := Conf.Get("mysql.pass")
	db_name := Conf.Get("mysql.db")
		
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", db_user, db_pass, db_name))
    if err != nil {
        log.Fatal(err)
    }
	defer db.Close()

    rows, err := db.Query("select * from employee")
    if err != nil {
        log.Fatal(err)
    }

	emps := []Employee{}
	for rows.Next() {
		var emp Employee
		err := rows.Scan(&emp.ID, &emp.Job_Title, &emp.Email_Address, &emp.FirstName_LastName)
		if err != nil {
			log.Fatal(err)
		}

		var emp_encr Employee
		//emp.ID, _ = Encrypt(emp.ID)
		emp_encr.ID = emp.ID
		emp_encr.Job_Title, _ = Encrypt(emp.Job_Title)		
		emp_encr.Email_Address, _ = Encrypt(emp.Email_Address)
		emp_encr.FirstName_LastName, _ = Encrypt(emp.FirstName_LastName)

		emps = append(emps, emp_encr)
	}

	c.JSON(200, emps)
}

func get_users(c *gin.Context) {
	
	fmt.Println("[GET USERS]")

	db_user := Conf.Get("mysql.user")
	db_pass := Conf.Get("mysql.pass")
	db_name := Conf.Get("mysql.db")
		
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", db_user, db_pass, db_name))
    if err != nil {
        log.Fatal(err)
    }
	defer db.Close()

    rows, err := db.Query("select * from employee")
    if err != nil {
        log.Fatal(err)
    }

	emps := []Employee{}
	for rows.Next() {
		var emp Employee
		err := rows.Scan(&emp.ID, &emp.Job_Title, &emp.Email_Address, &emp.FirstName_LastName)
		if err != nil {
			log.Fatal(err)
		}
		emps = append(emps, emp)
	}

	c.JSON(200, emps)
}

func home(c *gin.Context) {
	jsonData := []byte(`{"msg":"Welcome to BE Services", "services": ["/get_users", "/get_users_encr"]}`)	
	c.Data(200, "application/json", jsonData)
}

func init() {
		
	Conf = viper.New()
	Conf.SetConfigFile("config/default.json")	

	err := Conf.ReadInConfig()	
	if err != nil {
		fmt.Println(err)
	}

	CIPHER_KEY = Conf.Get("crypto.cipher_key").(string)
	if CIPHER_KEY == "" {
		fmt.Println("Cannot get cipher key!")
	}
}

func main() {
	
	r := gin.Default()

	r.GET("/", home)
	r.GET("/get_users", get_users)	
	r.GET("/get_users_encr", get_users_encr)	
	r.Run(":3000")
}
