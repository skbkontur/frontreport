package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	ch             *amqp.Channel
	port           = flag.String("port", "8888", "port to listen")
	amqpConnection = flag.String("amqp", "amqp://guest:guest@localhost:5672/", "amqp connection string")
	caCertPath     = flag.String("cacert", "", "custom CA cert file for SSL connections")
)

type CSPbody struct {
	DocumentURI       string `json:"document-uri"`
	Referrer          string `json:"referrer"`
	BlockedURI        string `json:"blocked-uri"`
	ViolatedDirective string `json:"violated-directive"`
	OriginalPolicy    string `json:"original-policy"`
	Timestamp         string `json:"@timestamp"`
}

type CSPreport struct {
	CSPbody *CSPbody `json:"csp-report"`
}

func report(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dec := json.NewDecoder(r.Body)
	bodyParsed := &CSPreport{}
	if err := dec.Decode(bodyParsed); err != nil {
		log.Printf("malformed JSON body: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	bodyParsed.CSPbody.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05.999Z")

	bodyDump, err := json.Marshal(bodyParsed.CSPbody)
	if err != nil {
		log.Printf("error dumping back JSON body: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	message := fmt.Sprintf(
		"{\"index\": {\"_index\": \"csp-report-%s\", \"_type\": \"csp-report\"}}\n%s\n",
		time.Now().Format("2006.01.02"),
		string(bodyDump))

	err = ch.Publish(
		"elasticsearch",
		"elasticsearch",
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			Timestamp:    time.Now(),
			Body:         []byte(message),
		})
	if err != nil {
		log.Printf("failed to send message to RabbitMQ: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	flag.Parse()

	var (
		conn *amqp.Connection
		err  error
	)
	if *caCertPath != "" {
		cfg := new(tls.Config)
		cfg.RootCAs = x509.NewCertPool()
		if ca, err := ioutil.ReadFile(*caCertPath); err == nil {
			cfg.RootCAs.AppendCertsFromPEM(ca)
		}
		conn, err = amqp.DialTLS(*amqpConnection, cfg)
	} else {
		conn, err = amqp.Dial(*amqpConnection)
	}
	if err != nil {
		log.Printf("error connecting to RabbitMQ: %s", err.Error())
		return
	}
	defer conn.Close()

	ch, err = conn.Channel()
	if err != nil {
		log.Printf("error creating channel: %s", err.Error())
		return
	}
	defer ch.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", report)
	http.ListenAndServe(fmt.Sprintf(":%s", *port), mux)
}
