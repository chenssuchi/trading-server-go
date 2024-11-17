package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"trade/config"
)

var PG *gorm.DB

func InitPg() {
	pgConf := config.Config.Pg
	db, err := gorm.Open(postgres.Open(pgConf.ConnStr), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	PG = db
	fmt.Println("Connected to database")
}
