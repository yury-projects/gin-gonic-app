package database

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	// Needed for gorm postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Database - returns database connection or panics
func Database() *gorm.DB {
	// Open a db connection
	db, err := gorm.Open("postgres", os.Getenv("GIN_GONIC_DATABASE_URL"))

	if err != nil {
		fmt.Println(err)
		panic("Failed to connect database")
	}

	return db
}
