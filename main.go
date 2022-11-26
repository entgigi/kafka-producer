package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	kafka "github.com/segmentio/kafka-go"
)

const (
	DefaultPort         = "3333"
	DefaultKafkaAddress = "localhost:9092"
	DefaultTopic        = "my-topic"
	dataset             = "{ title: \"mytitle\", body:\"my message\"}"
)

type config struct {
	Port         string
	KafkaAddress string
	Topic        string
}

func main() {
	log.Println("starting server ...")

	config := getConfigFromEnv()

	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(config.KafkaAddress),
		Topic:    config.Topic,
		Balancer: &kafka.RoundRobin{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/hello", getHello)
	mux.HandleFunc("/producer", handle_producer(kafkaWriter))

	// Catch the Ctrl-C and SIGTERM from kill command
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)

	go func() {
		signalType := <-ch
		signal.Stop(ch)
		log.Println("Exit command received. Exiting...")

		// this is a good place to flush everything to disk
		// before terminating.
		log.Println("Signal type : ", signalType)

		os.Exit(0)

	}()

	err := http.ListenAndServe(":"+config.Port, mux)

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}

}

func getConfigFromEnv() config {
	config := config{}
	config.Port = os.Getenv("PORT")
	if len(config.Port) <= 0 {
		config.Port = DefaultPort
	}
	config.KafkaAddress = os.Getenv("KAFKA")
	if len(config.KafkaAddress) <= 0 {
		config.KafkaAddress = DefaultKafkaAddress
	}
	config.Topic = os.Getenv("TOPIC")
	if len(config.Topic) <= 0 {
		config.Topic = DefaultTopic
	}
	return config
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	log.Println("got / request")
	io.WriteString(w, "This is my website!\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	log.Println("got /hello request")
	io.WriteString(w, "Hello, HTTP!\n")
}

func handle_producer(writer *kafka.Writer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("got /producer request")
		counter := r.PostFormValue("counter")
		n, err := strconv.Atoi(counter)
		if err != nil {
			log.Printf("Failed to convert count:%v", err)
			http.Error(w, "count value is not a valid number", http.StatusBadRequest)
			return
		}

		go write_messages(writer, n)
		io.WriteString(w, fmt.Sprintf("Producer, counter=%d\n", n))
	}
}

func write_messages(writer *kafka.Writer, count int) {
	log.Printf("start write message to Kafka")
	for i := 0; i < count; i++ {
		payload := []byte(dataset)
		err := writer.WriteMessages(context.Background(), kafka.Message{
			// Key:   []byte("key"),
			Value: payload,
		})
		if err != nil {
			log.Printf("Failed to write message:%v", err)
			continue
		}
	}
}
