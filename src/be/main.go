package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	// "errors"
	"encoding/json"
	"fmt"	
	"github.com/gin-gonic/gin"	
	"github.com/go-redis/redis"
	"github.com/mergermarket/go-pkcs7"	
	"github.com/spf13/viper"	
	"io"	
	"log"
	"strings"
	// "net/http"
	
    _ "github.com/go-sql-driver/mysql"	
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

type JsonType struct {
	Array []string
}

func get_employees_encr(c *gin.Context) {
		
	//cache will be based on an API Key. 
    //access without API key will be rejected
	key_recv := c.Query("api_key")	
	if key_recv == "" || len(key_recv) != 128 {
		c.Abort()		
		c.IndentedJSON(401, gin.H{"message":"Unauthorized Access"})
		return
	}
	
	key_decr, err := Decrypt(key_recv)
	if err != nil {
		c.Abort()
		c.IndentedJSON(401, gin.H{"message":"Unauthorized Access"})
		return;
	}

	api_keys := Conf.GetStringSlice("app.api_keys")

	containsKey := false
	for _, key := range api_keys {
		if key == key_decr {
			containsKey = true
			break;
		}
	}

	if !containsKey {
		c.Abort()
		c.IndentedJSON(401, gin.H{"message":"Unauthorized Access"})
		return;
	}

	//Redis operations caching, throttling, or db
	query := c.Request.URL.Query()

	avail_col := []string {"id","job_title", "email_address", "firstName_LastName"}
	contains_col := false

	qKey := ""
	qVal := ""

	for k, v := range query {
		if k == "api_key" {
			continue
		}
		if contains_col == false {
			for _, col := range avail_col {
				if col == k {
					contains_col = true
					qKey = k					
					qVal = strings.Join(v, " ")
				}
			}
		}				
	}

	// look at redis first, then database
	if qKey != "" && qVal != "" {
		
		client := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
			Password: "",
			DB: 0,
		})

		rKey := qKey + "=" + qVal		
		emps := []Employee{}

		//Check if key/value exists, if not, set value
		// vResult, err := client.Get(qKey).Result()
		result, err := client.Get(rKey).Result()		
		if err == redis.Nil {
			//if id, get data directly from database
			//if qKey == "id" {}
			fmt.Println("[DOES NOT CONTAIN DATA]")

			emps = employee_query("select * from employee", qKey, qVal)
			empsJson, err := json.Marshal(emps)
			if err != nil {
				c.Abort()
				c.IndentedJSON(500, gin.H{"message":"Server Error"})
				return;
			}
			
			if qKey == "id" {
				c.JSON(200, emps)
				return		
			}

			err = client.SAdd(rKey, empsJson).Err()
			if err != nil {
				c.Abort()
				c.IndentedJSON(500, gin.H{"message":"Server Error"})
				return;
			}
										
			c.JSON(200, emps)
			return		
		}		
		
		fmt.Println("[CONTAINs DATA]")
		err = json.Unmarshal([]byte(result), &emps)
		if err != nil {
			c.Abort()
			c.IndentedJSON(500, gin.H{"message":"Server Error"})
			return;
		}
		c.JSON(200, emps)
		return		
	}

	c.IndentedJSON(200, employee_query("select * from employee", "", ""))	
}

func employee_query(query, col, val string) ([]Employee) {
	
	db_user := Conf.Get("mysql.user")
	db_pass := Conf.Get("mysql.pass")
	db_name := Conf.Get("mysql.db")
		
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", db_user, db_pass, db_name))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	emps := []Employee{}	
	if col == "" && val == "" {
		rows, err := db.Query("select * from employee")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			var emp Employee
			err := rows.Scan(&emp.ID, &emp.Job_Title, &emp.Email_Address, &emp.FirstName_LastName)
			if err != nil {
				log.Fatal(err)
			}
	
			var emp_encr Employee		
			emp_encr.ID = emp.ID
			emp_encr.Job_Title, _ = Encrypt(emp.Job_Title)		
			emp_encr.Email_Address, _ = Encrypt(emp.Email_Address)
			emp_encr.FirstName_LastName, _ = Encrypt(emp.FirstName_LastName)
	
			emps = append(emps, emp_encr)
		}
	} else {						
		sql := fmt.Sprintf("select * from employee where %s like ?", col)
		rows, err := db.Query(sql, "%" + val + "%")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			var emp Employee
			err := rows.Scan(&emp.ID, &emp.Job_Title, &emp.Email_Address, &emp.FirstName_LastName)
			if err != nil {
				log.Fatal(err)
			}
	
			var emp_encr Employee		
			emp_encr.ID = emp.ID
			emp_encr.Job_Title, _ = Encrypt(emp.Job_Title)		
			emp_encr.Email_Address, _ = Encrypt(emp.Email_Address)
			emp_encr.FirstName_LastName, _ = Encrypt(emp.FirstName_LastName)
	
			emps = append(emps, emp_encr)
		}
	}
		
	return emps
}

func get_employees(c *gin.Context) {
	
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
	r.GET("/get_employees", get_employees)	
	r.GET("/get_employees_encr", get_employees_encr)	
	r.Run(":3000")
}
