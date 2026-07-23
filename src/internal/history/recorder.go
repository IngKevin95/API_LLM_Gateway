package history

import (
	"context"
	"log"
	"sync"
)

type Record struct {
	ID        string
	Model     string
	LatencyMs int64
	Success   bool
	Tokens    int
	Cost      float64
	Feedback  int
	Payload   []byte
}

type Persister interface {
	Save(ctx context.Context, rec Record) error
	UpdateFeedback(ctx context.Context, id string, feedback int) error
}

type Redactor interface {
	Redact(ctx context.Context, payload []byte) ([]byte, error)
}

type Recorder struct {
	persister Persister
	redactor  Redactor
	ch        chan Record
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

func NewRecorder(persister Persister, redactor Redactor, bufferSize int) *Recorder {
	return &Recorder{
		persister: persister,
		redactor:  redactor,
		ch:        make(chan Record, bufferSize),
		stopCh:    make(chan struct{}),
	}
}

func (r *Recorder) Start() {
	r.wg.Add(1)
	go r.worker()
}

func (r *Recorder) Stop() {
	close(r.stopCh)
	r.wg.Wait()
}

func (r *Recorder) Record(rec Record) {
	select {
	case r.ch <- rec:
	default:
		// Drop silently if full (or log metric)
		log.Printf("history recorder: queue full, dropping record %s", rec.ID)
	}
}

func (r *Recorder) AddFeedback(id string, feedback int) {
	// For simplicity, we just block here or send a special message.
	// Since UpdateFeedback might be an HTTP call from a client later,
	// doing it sync or via another worker is fine. Let's do it sync for now 
	// since it's a separate API endpoint, but it could be async too.
	// To make the test pass and be simple, we just call the persister.
	_ = r.persister.UpdateFeedback(context.Background(), id, feedback)
}

func (r *Recorder) worker() {
	defer r.wg.Done()
	for {
		select {
		case rec := <-r.ch:
			if r.redactor != nil && len(rec.Payload) > 0 {
				redacted, err := r.redactor.Redact(context.Background(), rec.Payload)
				if err == nil {
					rec.Payload = redacted
				}
			}
			err := r.persister.Save(context.Background(), rec)
			if err != nil {
				log.Printf("history recorder: save error: %v", err)
			}
		case <-r.stopCh:
			return
		}
	}
}
