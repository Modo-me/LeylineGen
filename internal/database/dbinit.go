package database

import (
	"fmt"
	"os"
	"quest_generator/internal/module/task"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DB_INIT() *gorm.DB {
	host := os.Getenv("DB_HOST")
	dsn := "user=username password=psw  host=%s port=5432 dbname=yourDbName sslmode=disable"
	dsn = fmt.Sprintf(dsn, host)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	err = dbSync(db)
	if err != nil {
		panic(err)
	}
	return db
}

func dbSync(db *gorm.DB) error {
	err := db.AutoMigrate(&task.Task{})
	if err != nil {
		return err
	}
	return nil
}
