package collect

import (
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
)

type HistoryMap struct {
	//Map   sync.Map
	Map *cache.Cache
	//Cache *cache.Cache
}

type CounterStats struct {
	Value float64
	Ts    int64
}

type CommonCounterHis struct {
	Snap [2]*CounterStats
	sync.RWMutex
}

func NewHistoryMap() *HistoryMap {
	//m := sync.Map{}
	c := cache.New(5*time.Minute, 10*time.Minute)
	return &HistoryMap{Map: c}

}

func NewCommonCounterHis() *CommonCounterHis {
	s := [2]*CounterStats{}
	return &CommonCounterHis{Snap: s}

}

func (cs *CommonCounterHis) UpdateCounterStat(this CounterStats) error {
	cs.Lock()
	defer cs.Unlock()
	for i := 1; i > 0; i-- {
		cs.Snap[i] = cs.Snap[i-1]
	}
	cs.Snap[0] = &this
	return nil
}

func (cs *CommonCounterHis) DeltaCounter() float64 {
	cs.Lock()
	defer cs.Unlock()
	if cs.Snap[1] == nil {
		return 0.0
	}
	valueDelta := cs.Snap[0].Value - cs.Snap[1].Value
	tsDelta := cs.Snap[0].Ts - cs.Snap[1].Ts
	// 0/0==NaN
	if tsDelta == 0 {
		return 0.0
	}

	return (valueDelta) / float64(tsDelta)

}
