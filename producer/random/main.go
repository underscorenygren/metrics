package random

import (
	"github.com/underscorenygren/producer"
	"math/rand"
	"time"
)

type random struct {
	failureRate float32
	p           producer.Producer
	rand        *math.Rand
}

//New a producer that fails messages randomly
func New(failureRate float32, p producer.Producer) producer.Producer {

	return &random{failureRate: failureRate,
		p:    p,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *random) PutRecords(records [][]byte) [][]byte {
	results := make([]byte, len(records))
	pruned := [][]byte{}
	for i, rec := range records {
		if r.randFloat32() < r.failureRate {
			results[i] = rec
		} else {
			pruned = append(pruned, rec)
			results[i] = nil
		}
	}

	real := r.PutRecords(pruned)
	j := 0
	for i := 0; i <= len(records); i++ {
		res := results[i]
		if res == nil {
			results[i] = real[j]
			j++
		}
	}

	return results
}
