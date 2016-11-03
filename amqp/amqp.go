package amqp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/assembla/cony"
	"github.com/facebookgo/muster"
	"github.com/streadway/amqp"
	"gopkg.in/tomb.v2"

	"github.com/skbkontur/cspreport"
)

// ReportStorage is an AMQP implementation of cspreport.ReportStorage interface
type ReportStorage struct {
	MaxBatchSize         uint
	MaxConcurrentBatches uint
	BatchTimeout         time.Duration
	PendingWorkCapacity  uint
	ExchangeName         string
	RoutingKey           string
	AMQPConnectionString string
	Logger               cspreport.Logger
	publisher            *cony.Publisher
	muster               muster.Client
	tomb                 tomb.Tomb
}

// Start initializes AMQP connections and muster batching
func (rs *ReportStorage) Start() error {
	client := cony.NewClient(
		cony.URL(rs.AMQPConnectionString),
		cony.Backoff(cony.DefaultBackoff),
	)

	exchange := cony.Exchange{
		Name:    rs.ExchangeName,
		Kind:    "direct",
		Durable: true,
	}
	client.Declare([]cony.Declaration{
		cony.DeclareExchange(exchange),
	})

	rs.publisher = cony.NewPublisher(rs.ExchangeName, rs.RoutingKey)
	client.Publish(rs.publisher)

	rs.tomb.Go(func() error {
		for client.Loop() {
			select {
			case <-rs.tomb.Dying():
				client.Close()
			case err := <-client.Errors():
				rs.Logger.Log("msg", "error communicating with remote server", "error", err)
			}
		}
		return nil
	})

	rs.muster.MaxBatchSize = rs.MaxBatchSize
	rs.muster.MaxConcurrentBatches = rs.MaxConcurrentBatches
	rs.muster.BatchTimeout = rs.BatchTimeout
	rs.muster.PendingWorkCapacity = rs.PendingWorkCapacity
	rs.muster.BatchMaker = func() muster.Batch { return &batch{ReportStorage: rs} }

	err := rs.muster.Start()
	return err
}

// Stop flushes and stops muster batching
func (rs *ReportStorage) Stop() error {
	rs.tomb.Go(func() error {
		timer := time.NewTimer(10 * time.Second)
		select {
		case <-rs.tomb.Dying():
			return nil
		case <-timer.C:
			rs.publisher.Cancel()
			return errors.New("at least one publishing timed out, had to cancel")
		}
	})

	errMuster := rs.muster.Stop()

	rs.tomb.Kill(nil)
	errTomb := rs.tomb.Wait()

	if errMuster != nil {
		return errMuster
	}
	return errTomb
}

// AddCSPReport adds a report to next batch
func (rs *ReportStorage) AddCSPReport(report cspreport.CSPReport) {
	decoratedReport := bytes.NewBufferString(fmt.Sprintf("{\"index\": {\"_index\": \"csp-report-%s\", \"_type\": \"csp-report\"}}\n", time.Now().UTC().Format("2006.01.02")))
	encoder := json.NewEncoder(decoratedReport)
	if err := encoder.Encode(&report); err != nil {
		rs.Logger.Log("msg", "failed to encode", "report_type", "CSP", "error", err)
	} else {
		decoratedReport.WriteString("\n")
		rs.muster.Work <- decoratedReport.Bytes()
	}
}

// AddPKPReport adds a report to next batch
func (rs *ReportStorage) AddPKPReport(report cspreport.PKPReport) {
	decoratedReport := bytes.NewBufferString(fmt.Sprintf("{\"index\": {\"_index\": \"pkp-report-%s\", \"_type\": \"pkp-report\"}}\n", time.Now().UTC().Format("2006.01.02")))
	encoder := json.NewEncoder(decoratedReport)
	if err := encoder.Encode(&report); err != nil {
		rs.Logger.Log("msg", "failed to encode", "report_type", "PKP", "error", err)
	} else {
		decoratedReport.WriteString("\n")
		rs.muster.Work <- decoratedReport.Bytes()
	}
}

type batch struct {
	ReportStorage *ReportStorage
	Items         bytes.Buffer
}

func (b *batch) Add(item interface{}) {
	b.Items.Write(item.([]byte))
}

func (b *batch) Fire(notifier muster.Notifier) {
	defer notifier.Done()
	err := b.ReportStorage.publisher.Publish(
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			Timestamp:    time.Now(),
			Body:         b.Items.Bytes(),
		})
	if err != nil {
		b.ReportStorage.Logger.Log("msg", "failed to fire batch", "size", b.Items.Len(), "error", err)
	}
}
