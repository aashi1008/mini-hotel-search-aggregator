package providers

import (
	"math/rand"
	"time"
)

func sampleLatencyFromRng(rng *rand.Rand, avg float64) time.Duration {
	ms := float64(50) + rng.ExpFloat64()*avg*200.0
	return time.Duration(ms) * time.Millisecond
}

func shouldFailFromRng(rng *rand.Rand, rate float64) bool {
	return rng.Float64() < rate
}
