package collect

import (
	"sync"
)

type HistoryMap struct {
	Map sync.Map
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
	m := sync.Map{}
	return &HistoryMap{Map: m}

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
	if valueDelta == 0 && tsDelta == 0 {
		return 0.0
	}

	return (valueDelta) / float64(tsDelta)

}
