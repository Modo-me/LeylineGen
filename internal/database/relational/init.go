package relational

import (
	"fmt"
	"strings"

	"quest_generator/internal/module/task"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DbInit() *gorm.DB {
	initConfig()

	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		viper.GetString("relational_database.user"),
		viper.GetString("relational_database.password"),
		viper.GetString("relational_database.host"),
		viper.GetInt("relational_database.port"),
		viper.GetString("relational_database.dbname"),
		viper.GetString("relational_database.sslmode"),
	)

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
	return db.AutoMigrate(&task.Task{})
}

func initConfig() {
	// Defaults match the original hardcoded DSN values
	viper.SetDefault("relational_database.user", "username")
	viper.SetDefault("relational_database.password", "psw")
	viper.SetDefault("relational_database.host", "localhost")
	viper.SetDefault("relational_database.port", 5432)
	viper.SetDefault("relational_database.dbname", "yourDbName")
	viper.SetDefault("relational_database.sslmode", "disable")

	// config.yaml (optional) overrides defaults
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	_ = viper.ReadInConfig() // ignore missing file

	// Env var DB_HOST overrides database.host (backward compat)
	_ = viper.BindEnv("database.host", "DB_HOST")

	// All database.* fields can also be overridden by DATABASE_* env vars
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}
