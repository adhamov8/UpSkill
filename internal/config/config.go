package config

import "os"

type Config struct {
	PgConnStr string
	JwtSecret []byte
	KafkaAddr string
	TopicName string
}

func NewConfig() *Config {
	pg := os.Getenv("PG_CONN_STR")
	if pg == "" {
		pg = "host=localhost port=5432 user=postgres password=postgres dbname=upskill sslmode=disable"
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "SUPER_SECRET_KEY"
	}
	kafkaAddr := os.Getenv("KAFKA_ADDR")
	if kafkaAddr == "" {
		kafkaAddr = "localhost:9092"
	}
	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "upskill_events"
	}

	return &Config{
		PgConnStr: pg,
		JwtSecret: []byte(secret),
		KafkaAddr: kafkaAddr,
		TopicName: topic,
	}
}
