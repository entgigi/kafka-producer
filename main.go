package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	DefaultPort = "3333"
)

func main() {
	log.Println("starting server ...")

	mux := http.NewServeMux()
	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/hello", getHello)
	mux.HandleFunc("/producer", postProducer)

	port := os.Getenv("PORT")
	if len(port) <= 0 {
		port = DefaultPort
	}

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

	err := http.ListenAndServe(":"+port, mux)

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}

}

func getRoot(w http.ResponseWriter, r *http.Request) {
	log.Println("got / request")
	io.WriteString(w, "This is my website!\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	log.Println("got /hello request")
	io.WriteString(w, "Hello, HTTP!\n")
}

func postProducer(w http.ResponseWriter, r *http.Request) {
	log.Println("got /producer request")
	counter := r.PostFormValue("counter")
	io.WriteString(w, fmt.Sprintf("Producer, counter=%s\n", counter))
}
