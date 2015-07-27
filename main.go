package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/streadway/amqp"
)

var (
	ch             *amqp.Channel
	port           = flag.String("port", "8888", "port to listen")
	amqpConnection = flag.String("amqp", "amqp://guest:guest@localhost:5672/", "amqp connection string")
	caCertPath     = flag.String("cacert", "", "custom CA cert file for SSL connections")
)

type cspBody struct {
	DocumentURI       string `json:"document-uri"`
	Referrer          string `json:"referrer"`
	BlockedURI        string `json:"blocked-uri"`
	ViolatedDirective string `json:"violated-directive"`
	OriginalPolicy    string `json:"original-policy"`
	Timestamp         string `json:"@timestamp"`
}

type cspReport struct {
	Body *cspBody `json:"csp-report"`
}

type pkpReport struct {
	DateTime                  string   `json:"date-time"`
	Hostname                  string   `json:"hostname"`
	Port                      int      `json:"port"`
	EffectiveExpirationDate   string   `json:"effective-expiration-date"`
	IncludeSubdomains         bool     `json:"include-subdomains"`
	NotedHostname             string   `json:"noted-hostname"`
	ServedCertificateChain    []string `json:"served-certificate-chain"`
	ValidatedCertificateChain []string `json:"validated-certificate-chain"`
	KnownPins                 []string `json:"known-pins"`
}

func handleCspReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dec := json.NewDecoder(r.Body)
	bodyParsed := &cspReport{}
	if err := dec.Decode(bodyParsed); err != nil {
		log.Printf("malformed JSON body: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	bodyParsed.Body.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05.999Z")

	bodyDump, err := json.Marshal(bodyParsed.Body)
	if err != nil {
		log.Printf("error dumping back JSON body: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := sendParsedReport(bodyDump, "csp"); err != nil {
		log.Printf("failed to send message to RabbitMQ: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handlePkpReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dec := json.NewDecoder(r.Body)
	bodyParsed := &pkpReport{}
	if err := dec.Decode(bodyParsed); err != nil {
		log.Printf("malformed JSON body: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bodyDump, err := json.Marshal(bodyParsed)
	if err != nil {
		log.Printf("error dumping back JSON body: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := sendParsedReport(bodyDump, "pkp"); err != nil {
		log.Printf("failed to send message to RabbitMQ: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func sendParsedReport(reportBody []byte, reportType string) error {
	message := fmt.Sprintf(
		"{\"index\": {\"_index\": \"%s-report-%s\", \"_type\": \"%s-report\"}}\n%s\n",
		reportType,
		time.Now().Format("2006.01.02"),
		reportType,
		string(reportBody))

	err := ch.Publish(
		"csp",
		"csp",
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			Timestamp:    time.Now(),
			Body:         []byte(message),
		})
	return err
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
	mux.HandleFunc("/", handleCspReport)
	mux.HandleFunc("/csp", handleCspReport)
	mux.HandleFunc("/pkp", handlePkpReport)
	http.ListenAndServe(fmt.Sprintf(":%s", *port), mux)
}
