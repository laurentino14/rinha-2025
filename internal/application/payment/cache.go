package payment

import (
	"sync"
	"time"
)

type cache struct {
	M               sync.RWMutex
	ActiveProcessor int
	LastChecked     time.Time
}

func NewCache() *cache {
	return &cache{
		ActiveProcessor: 1,
		LastChecked:     time.Now(),
	}
}
