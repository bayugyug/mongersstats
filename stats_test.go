package mongersstats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatsInit(t *testing.T) {

	q, err := NewQ()
	if !assert.NoError(t, err, "New should succeed") {
		return
	}

	expected := 100
	i := 0
	for {
		i++
		q.Incr("TEST::STATS")
		q.FloatIncr("DECIMAL::STATS::TEST")
		if i >= 100 {
			break
		}

	}
	if !assert.Equal(t, expected, i) {
		return
	}
}
