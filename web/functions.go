package web

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"forum/models"
	"log"
	"strings"
	"time"
)

func GetCurrentTime() (time.Time, error) {
	timeNow := time.Now()
	modTime := timeNow.Format(("15:04:05 02 Jan 2006"))

	createdAt, err := time.Parse("15:04:05 02 Jan 2006", modTime)

	if err != nil {
		log.Println("Error parsing time:", err)
	}
	return createdAt, nil
}

func ClearData(tableName string) {
	query3 := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
	_, err := models.Db.Exec(query3)
	if err != nil {
		log.Println("Error occured during clearing sql data", err)
		return
	}
}

func GenerateSessionToken() (string, error) {
	token := make([]byte, 32) // 32 bytes = 256 bits
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(token), nil
}

func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	} else {
		return strings.ToUpper(string(s[0])) + s[1:]
	}
}
