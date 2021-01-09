package events

import (
	"fmt"
	"time"

	diskqueue "github.com/nsqio/go-diskqueue"
	"go.vocdoni.io/dvote/crypto/ethereum"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/vochain"
	"go.vocdoni.io/proto/build/go/models"
	"google.golang.org/protobuf/proto"
)

// EventWorker is responsible for performing a unit of work given an event
type EventWorker struct {
	Storage diskqueue.Interface
	// Signer
	Signer *ethereum.SignKeys
	// VochainApp is a pointer to the Vochain BaseApplication allowing to call SendTx method
	VochainApp *vochain.BaseApplication
}

// Start starts the event worker.
// The worker will constantly pull for new events
// and will handle them one by one
// Events can be handled sync or async and it is up
// to the callback to define the behaviour.
// i.e if we want OnProcessResults event to be processed
// async the implementation of the handler must be wrapped with
// an anonymous goroutine.
func (ew *EventWorker) Start() {
	for {
		// Receive a event.
		event, err := ew.NextEvent()
		if err != nil {
			log.Warnf("cannot get next event: %s", err)
			continue
		}
		if event == nil {
			time.Sleep(time.Millisecond * 200)
			continue
			// TODO: mvdan improve this
		}
		log.Debugf("processing event: %+v", event)
		switch event.SourceType {
		case ScrutinizerOnprocessresult:
			if err := ew.onComputeResults(event.Data.(*models.ProcessResult)); err != nil {
				log.Warnf("event processing failed: %s", err)
				continue
			}
		default:
			log.Warnf("cannot process event, invalid source type: %x", event.SourceType)
			continue
		}
		log.Infof("event %+v processed successfully", event)
	}
}

// OnComputeResults is called once a process result is computed by the scrutinizer.
func (ew *EventWorker) onComputeResults(results *models.ProcessResult) error {
	// create setProcessTx
	setprocessTxArgs := &models.SetProcessTx{
		ProcessId: results.ProcessId,
		Results:   results,
		Status:    models.ProcessStatus_RESULTS.Enum(),
		Txtype:    models.TxType_SET_PROCESS_RESULTS,
	}

	vtx := models.Tx{}
	resultsTxBytes, err := proto.Marshal(setprocessTxArgs)
	if err != nil {
		return fmt.Errorf("cannot marshal set process results tx: %w", err)
	}
	vtx.Signature, err = ew.Signer.Sign(resultsTxBytes)
	if err != nil {
		return fmt.Errorf("cannot sign oracle tx: %w", err)
	}

	vtx.Payload = &models.Tx_SetProcess{SetProcess: setprocessTxArgs}
	txb, err := proto.Marshal(&vtx)
	if err != nil {
		return fmt.Errorf("error marshaling set process results tx: %w", err)
	}
	log.Debugf("broadcasting Vochain Tx: %s", setprocessTxArgs.String())

	res, err := ew.VochainApp.SendTX(txb)
	if err != nil || res == nil {
		return fmt.Errorf("cannot broadcast tx: %w, res: %+v", err, res)
	}
	log.Infof("transaction sent, hash:%s", res.Hash)
	return nil

}

// NextEvent gets an event from the events persistent storage
func (ew *EventWorker) NextEvent() (*Event, error) {
	event := &Event{}
	ev := <-ew.Storage.ReadChan()
	switch ev[0] {
	case ScrutinizerOnprocessresult:
		evData := &models.ProcessResult{}
		if err := proto.Unmarshal(ev[1:], evData); err != nil {
			return nil, err
		}
		event.Data = evData
	}
	event.SourceType = ev[0]
	return event, nil
}
