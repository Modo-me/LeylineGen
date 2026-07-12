package graph

import (
	"log"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/spf13/viper"
)

func DbInit() neo4j.Driver {
	initConfig()

	dbUri := viper.GetString("graph_database.uri")
	dbUser := viper.GetString("graph_database.user")
	dbPassword := viper.GetString("graph_database.password")

	driver, err := neo4j.NewDriver(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""),
	)
	if err != nil {
		panic(err)
	}

	log.Printf("Connected to %s", dbUri)
	return driver
}

func initConfig() {
	// Defaults match the original hardcoded values
	viper.SetDefault("graph_database.uri", "neo4j://localhost:7687")
	viper.SetDefault("graph_database.user", "neo4j")
	viper.SetDefault("graph_database.password", "neo4j")

	// config.yaml (optional) overrides defaults
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	_ = viper.ReadInConfig() // ignore missing file

	// All graph.* fields can also be overridden by GRAPH_* env vars
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}
