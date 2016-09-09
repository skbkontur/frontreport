package amqp

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"time"

	"github.com/facebookgo/muster"
	"github.com/skbkontur/cspreport"
	"github.com/streadway/amqp"
)

// ReportStorage is an AMQP implementation of cspreport.ReportStorage interface
type ReportStorage struct {
	MaxBatchSize         uint
	BatchTimeout         time.Duration
	PendingWorkCapacity  uint
	Exchange             string
	RoutingKey           string
	AMQPConnectionString string
	CACert               []byte
	Logger               cspreport.Logger
	connection           *amqp.Connection
	musterCSP            muster.Client
	musterPKP            muster.Client
}

// Start initializes muster batching
func (rs *ReportStorage) Start() error {
	if err := rs.dialAMQP(); err != nil {
		return err
	}

	rs.musterCSP.MaxBatchSize = rs.MaxBatchSize
	rs.musterPKP.MaxBatchSize = rs.MaxBatchSize
	rs.musterCSP.BatchTimeout = rs.BatchTimeout
	rs.musterPKP.BatchTimeout = rs.BatchTimeout
	rs.musterCSP.PendingWorkCapacity = rs.PendingWorkCapacity
	rs.musterPKP.PendingWorkCapacity = rs.PendingWorkCapacity
	rs.musterCSP.BatchMaker = func() muster.Batch { return &batchCSP{Storage: rs} }
	rs.musterPKP.BatchMaker = func() muster.Batch { return &batchPKP{Storage: rs} }
	errCSP := rs.musterCSP.Start()
	errPKP := rs.musterPKP.Start()
	if errCSP != nil {
		return errCSP
	}
	return errPKP
}

// Stop flushes and stops muster batching
func (rs *ReportStorage) Stop() error {
	errCSP := rs.musterCSP.Stop()
	errPKP := rs.musterPKP.Stop()
	errCloseConnection := rs.connection.Close()
	if errCSP != nil {
		return errCSP
	}
	if errPKP != nil {
		return errPKP
	}
	return errCloseConnection
}

// AddCSPReport adds a report to next batch
func (rs *ReportStorage) AddCSPReport(report cspreport.CSPReport) {
	rs.musterCSP.Work <- report
}

// AddPKPReport adds a report to next batch
func (rs *ReportStorage) AddPKPReport(report cspreport.PKPReport) {
	rs.musterPKP.Work <- report
}

func (rs *ReportStorage) dialAMQP() error {
	var err error
	if len(rs.CACert) > 0 {
		cfg := new(tls.Config)
		cfg.RootCAs = x509.NewCertPool()
		cfg.RootCAs.AppendCertsFromPEM(rs.CACert)
		rs.connection, err = amqp.DialTLS(rs.AMQPConnectionString, cfg)
	} else {
		rs.connection, err = amqp.Dial(rs.AMQPConnectionString)
	}
	return err
}

func (rs *ReportStorage) sendBatchToAMQP(batch []byte) error {
	channel, err := rs.connection.Channel()
	if err != nil {
		return err
	}
	return channel.Publish(
		rs.Exchange,
		rs.RoutingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			Timestamp:    time.Now(),
			Body:         batch,
		})
}

type batchCSP struct {
	Storage *ReportStorage
	Items   []cspreport.CSPReport
}

func (b *batchCSP) Add(item interface{}) {
	b.Items = append(b.Items, item.(cspreport.CSPReport))
}

func (b *batchCSP) Fire(notifier muster.Notifier) {
	defer notifier.Done()
	batch := bytes.NewBufferString(fmt.Sprintf("{\"index\": {\"_index\": \"csp-report-%s\", \"_type\": \"csp-report\"}}\n", time.Now().UTC().Format("2006.01.02")))
	encoder := json.NewEncoder(batch)
	for _, item := range b.Items {
		if err := encoder.Encode(&item.Body); err != nil {
			b.Storage.Logger.Log("msg", "failed to encode", "report_type", "CSP", "error", err)
		}
		batch.WriteString("\n")
	}
	if len(b.Items) > 0 {
		b.Storage.Logger.Log("msg", "sent non-empty batch", "report_type", "CSP", "count", len(b.Items))
	}
	if err := b.Storage.sendBatchToAMQP(batch.Bytes()); err != nil {
		b.Storage.Logger.Log("msg", "failed to send batch", "report_type", "CSP", "error", err)
	}
}

type batchPKP struct {
	Storage *ReportStorage
	Items   []cspreport.PKPReport
}

func (b *batchPKP) Add(item interface{}) {
	b.Items = append(b.Items, item.(cspreport.PKPReport))
}

func (b *batchPKP) Fire(notifier muster.Notifier) {
	defer notifier.Done()
	batch := bytes.NewBufferString(fmt.Sprintf("{\"index\": {\"_index\": \"pkp-report-%s\", \"_type\": \"pkp-report\"}}\n", time.Now().UTC().Format("2006.01.02")))
	encoder := json.NewEncoder(batch)
	for _, item := range b.Items {
		if err := encoder.Encode(&item); err != nil {
			b.Storage.Logger.Log("msg", "failed to encode", "report_type", "PKP", "error", err)
		}
		batch.WriteString("\n")
	}
	b.Storage.Logger.Log("msg", "sent non-empty batch", "report_type", "PKP", "count", len(b.Items))
	if err := b.Storage.sendBatchToAMQP(batch.Bytes()); err != nil {
		b.Storage.Logger.Log("msg", "failed to send batch", "report_type", "PKP", "error", err)
	}
}
