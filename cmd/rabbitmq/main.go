package main

import (
	"encoding/json"
	"log"
	"net/smtp"

	"github.com/oogway93/taskmanager/config"
	"github.com/oogway93/taskmanager/internal/entity"
	"github.com/oogway93/taskmanager/logger"
	"github.com/streadway/amqp"
)

func main() {
	cfg := config.Load()

	Log := logger.Init(cfg)
	defer logger.Sync(Log)
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Ошибка подключения: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Ошибка открытия канала: %s", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"email_greetings",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Ошибка объявления очереди: %s", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Ошибка регистрации потребителя: %s", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var message entity.EmailMessage
			err := json.Unmarshal(d.Body, &message)
			if err != nil {
				log.Printf("Ошибка декодирования сообщения: %s", err)
				continue
			}

			err = sendEmail(cfg.Email.EmailFrom, cfg.Email.EmailPass, message.EmailTo)
			if err != nil {
				log.Printf("Ошибка отправки email: %s", err)
			} else {
				log.Printf("Email отправлен для: %s", message)
			}
		}
	}()

	log.Printf("Ожидание сообщений...")
	<-forever
}

func sendEmail(emailFrom, pass, emailTo string) error {

	msg := "From: " + emailFrom + "\n" +
		"To: " + emailTo + "\n" +
		"Subject: Hello there\n\n" +
		"Greetings my friend!!!"

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", emailFrom, pass, "smtp.gmail.com"),
		emailFrom, []string{emailTo}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return err
	}
	log.Println("Successfully sended to " + emailTo)
	return nil
}
