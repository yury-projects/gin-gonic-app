package database

import (
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func Database() *gorm.DB {
	// Open a db connection
	db, err := gorm.Open("postgres", os.Getenv("GIN_GONIC_DATABASE_URL"))

	if err != nil {
		panic("Failed to connect database")
	}

	return db
}
