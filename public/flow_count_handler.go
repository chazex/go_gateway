package public

import (
	"sync"
	"time"
)

var FlowCounterHandler *FlowCounter

type FlowCounter struct {
	RedisFlowCountMap   map[string]*RedisFlowCountService //map
	RedisFlowCountSlice []*RedisFlowCountService //slice ，有map为啥还要一个slice呢
	Locker              sync.RWMutex
}

func NewFlowCounter() *FlowCounter {
	return &FlowCounter{
		RedisFlowCountMap:   map[string]*RedisFlowCountService{},
		RedisFlowCountSlice: []*RedisFlowCountService{},
		Locker:              sync.RWMutex{},
	}
}

func init() {
	FlowCounterHandler = NewFlowCounter()
}

func (counter *FlowCounter) GetCounter(serverName string) (*RedisFlowCountService, error) {
	for _, item := range counter.RedisFlowCountSlice {
		if item.AppID == serverName {
			return item, nil
		}
	}

	// 没找到，新初始化一个
	newCounter:=NewRedisFlowCountService(serverName,1*time.Second)
	counter.RedisFlowCountSlice = append(counter.RedisFlowCountSlice, newCounter)
	
	// map并发写，需要加锁。
	counter.Locker.Lock()
	defer counter.Locker.Unlock()
	counter.RedisFlowCountMap[serverName] = newCounter
	return newCounter, nil
}
