package events

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func NewKafkaWriter(kafkaAddr, topicName string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(kafkaAddr),
		Topic:        topicName,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: 1,
	}
}

func ProduceEvent(writer *kafka.Writer, eventName, payload string) {
	msgValue := fmt.Sprintf("%s => %s", eventName, payload)
	msg := kafka.Message{
		Key:   []byte(eventName),
		Value: []byte(msgValue),
		Time:  time.Now(),
	}
	err := writer.WriteMessages(context.Background(), msg)
	if err != nil {
		log.Printf("[Kafka] Ошибка отправки события '%s': %v\n", msgValue, err)
	} else {
		log.Printf("[Kafka] Событие '%s' отправлено.\n", msgValue)
	}
}
