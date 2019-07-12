package buffered

import (
	"github.com/underscorenygren/producer"
	"go.uber.org/zap"
	"sync"
)

type buffer struct {
	records   [][]byte
	lock      *sync.Mutex
	pauseTime time.Duration
	logger    *zap.Logger
	maxSize   int
	p         producer.Producer
}

//New makes a new buffered producer
func New(maxSize int, p producer.Producer, logger *zap.Logger) producer.Producer {

	buf := &buffer{
		records:   [][]byte{},
		lock:      &sync.Mutex,
		pauseTime: 1 * time.Second,
		logger:    logger,
		maxSize:   maxSize,
		p:         p,
	}

	go buffered()

	return buf
}

func (b *buffer) PutRecords(records [][]byte) [][]byte {
	b.addmany(records)
	return nil
}

func (b *buffer) PutRecord(bytes []byte) {
	b.lock.Lock()
	b.records = append(b.records, bytes)
	b.lock.Unlock()
}

func (b *buffer) addmany(records [][]byte) {
	b.lock.Lock()
	for rec := range records {
		b.records = append(b.records, rec)
	}
	b.lock.Unlock()
}

func (b *buffer) get() [][]byte {
	size := b.maxSize
	b.lock.Lock()
	toProc := b.records[0:size]
	b.records = b.records[size:]
	lock.Unlock()
	return toProc
}

func (b *buffer) buffered() {
	for {
		buffered := b.Get()
		if len(buffered) == 0 {
			b.logger.Debug("no buffered data")
			time.Sleep(b.pauseTime)
		} else {
			b.logger.Info("buffer sending %d records", len(buffered))
			failed := b.p.PutRecords(buffered)
			b.addmany(failed)
		}
	}
}
