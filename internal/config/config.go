package config

import (
	"os"
)

type Config struct {
	PgConnStr  string
	JwtSecret  []byte
	KafkaAddr  string
	KafkaTopic string
}

func GetConfig() *Config {
	pg := getEnv("PG_CONN_STR", "host=localhost port=5432 user=postgres password=postgres dbname=upskill_db sslmode=disable")
	secret := []byte(getEnv("JWT_SECRET", "SUPER_SECRET_KEY"))
	kafkaAddr := getEnv("KAFKA_ADDR", "localhost:9092")
	topic := getEnv("KAFKA_TOPIC", "upskill_events")

	return &Config{
		PgConnStr:  pg,
		JwtSecret:  secret,
		KafkaAddr:  kafkaAddr,
		KafkaTopic: topic,
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
