package events

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func InitKafkaWriter(kafkaAddr, topic string) *kafka.Writer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(kafkaAddr),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: 1,
	}
	return writer
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
