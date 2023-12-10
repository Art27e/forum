package web

import (
	"time"
	"fmt"
	"log"
	"forum/models"
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
