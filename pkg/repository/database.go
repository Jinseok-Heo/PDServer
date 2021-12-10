package repository

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func NewDatabase() (*gorm.DB, error) {
	dbUserName := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		dbUserName, dbPassword, dbHost, dbPort, dbName)
	db, err := gorm.Open("mysql", connectionString)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if err := db.DB().Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// Close - Close database
func CloseDB(db *gorm.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
