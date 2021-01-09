package events

import (
	"fmt"
	"time"

	"github.com/nsqio/go-diskqueue"
	"go.vocdoni.io/dvote/crypto/ethereum"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
	"go.vocdoni.io/dvote/vochain"
	"go.vocdoni.io/proto/build/go/models"
	"google.golang.org/protobuf/proto"
)

const (
	// scrutinizer can use from 0 to 20 (excluded)

	// ScrutinizerOnprocessresult represents the event created once the scrutinizer computed successfully the results of a process
	ScrutinizerOnprocessresult = byte(0)

	// keykeeper can use from 20 to 40 (excluded)

	// ethereum can use from 40 to 60 (excluded)

	// vochain can use from 60 to 80 (excluded)

	// census can use from 80 to 100 (excluded)
)

// Event represents all the info required for an event to be processed
type Event struct {
	Data       proto.Message
	SourceType byte
}

// Dispatcher responsible for pulling work requests
// and distributing them to the next available worker
type Dispatcher struct {
	Signer *ethereum.SignKeys
	// Storage stores the unprocessed events
	Storage [len(types.EventSources)]diskqueue.Interface
	// VochainApp is a pointer to the Vochain BaseApplication allowing to call SendTx method
	VochainApp *vochain.BaseApplication
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(signer *ethereum.SignKeys, vochain *vochain.BaseApplication, dbPath string, maxBytesPerFile int64, minMsgSize, maxMsgSize int32, syncEvery int64, syncTimeout time.Duration, logger diskqueue.AppLogFunc) *Dispatcher {
	d := &Dispatcher{Signer: signer, VochainApp: vochain}
	for idx := range d.Storage {
		d.Storage[idx] = diskqueue.New(types.EventSources[idx], dbPath, maxBytesPerFile, minMsgSize, maxMsgSize, syncEvery, syncTimeout, logger)
	}
	return d
}

// StartDispatcher creates and starts the workers
func (d *Dispatcher) StartDispatcher() {
	// create all of our event workers.
	for i := 0; i < len(types.EventSources); i++ {
		log.Debugf("Starting event worker", i+1)
		eventWorker := &EventWorker{
			Storage:    d.Storage[i],
			Signer:     d.Signer,
			VochainApp: d.VochainApp,
		}
		go eventWorker.Start()
	}
	log.Infof("events dispatcher started with %d workers", len(types.EventSources))
}

// Collect receives, check and push the events to the event queue
func (d *Dispatcher) Collect(eventData proto.Message, sourceType byte) {
	// Now, we take the delay, and the person's name, and make a WorkRequest out of them.
	if sourceType < byte(0) || sourceType > byte(255) {
		log.Debugf("invalid source type: %d", sourceType)
	}
	// receive the event
	event, err := d.Wrap(eventData, sourceType)
	if err != nil {
		log.Debugf("cannot receive event: %w", err)
		return
	}

	// add event to the database
	if err := d.Add(event); err != nil {
		log.Debugf("cannot add event into database: %w", err)
		return
	}

	log.Infof("event %+v added for processing", event)
}

// Wrap checks and wraps an event into an Event struct
func (d *Dispatcher) Wrap(eventData proto.Message, sourceType byte) (*Event, error) {
	event := &Event{}
	switch sourceType {
	case ScrutinizerOnprocessresult:
		// check event data
		if err := checkResults(eventData.(*models.ProcessResult)); err != nil {
			return nil, fmt.Errorf("invalid event received: %w", err)
		}
	default:
		return nil, fmt.Errorf("cannot receive event, invalid event origin")
	}
	event.Data = eventData
	event.SourceType = sourceType
	return event, nil
}

func checkResults(results *models.ProcessResult) error {
	if len(results.EntityId) != types.EntityIDsize {
		return fmt.Errorf("invalid entityId size")
	}
	if len(results.ProcessId) != types.ProcessIDsize {
		return fmt.Errorf("invalid processId size")
	}
	if len(results.Votes) <= 0 {
		return fmt.Errorf("invalid votes length, results.Votes cannot be empty")
	}
	return nil
}

// DB

// Add adds an event to the events persistent storage.
// An event will be stored as: []byte where byte[0] encodes the event type.
// The other event bytes encode the event data itself
func (d *Dispatcher) Add(event *Event) error {
	var eventDataBytes []byte
	var err error
	switch event.SourceType {
	case ScrutinizerOnprocessresult:
		switch t := event.Data.(type) {
		case *models.ProcessResult:
			eventDataBytes, err = proto.Marshal(t)
			if err != nil {
				return fmt.Errorf("cannot marshal event %+v data: %w", event, err)
			}
		default:
			return fmt.Errorf("cannot add scrutinizer event to persisten storage, invalid event type")
		}
	default:
		return fmt.Errorf("cannot add event %+v to persistent storage, unsupported event origin: %d", event, event.SourceType)
	}
	queueIdx := 0
	switch est := event.SourceType; {
	case est < 20:
		// nothing to do, 0 is fine
	case est >= 20 && est < 40:
		queueIdx = 1
	case est >= 40 && est < 60:
		queueIdx = 2
	case est >= 60 && est < 80:
		queueIdx = 3
	case est >= 80 && est < 100:
		queueIdx = 4
	default:
		return fmt.Errorf("cannot add event, invalid queue index")
	}
	// byte[0]  -> event type
	// byte[1:] -> event data
	err = d.Storage[queueIdx].Put(append([]byte{event.SourceType}, eventDataBytes...))
	if err != nil {
		return fmt.Errorf("cannot add event %+v to persistent storage: %w", event, err)
	}
	return nil
}
