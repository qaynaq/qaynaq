package coordinator

import (
	"sync"
)

type FlowWorkerMap interface {
	GetFlowWorker(flowID int64) (string, bool)
	GetFlowWorkerStream(flowID int64) (int64, bool)
	SetFlowWorker(flowID int64, workerID string, workerFlowID int64)
	RemoveFlow(flowID int64)
	RemoveFlowIfMatches(flowID int64, workerFlowID int64)
}

type flowWorkerMap struct {
	mu                 sync.RWMutex
	streamWorker       map[int64]string
	streamWorkerStream map[int64]int64
}

func NewFlowWorkerMap() FlowWorkerMap {
	return &flowWorkerMap{
		streamWorker:       make(map[int64]string),
		streamWorkerStream: make(map[int64]int64),
	}
}

func (m *flowWorkerMap) GetFlowWorker(flowID int64) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	workerID, exists := m.streamWorker[flowID]
	return workerID, exists
}

func (m *flowWorkerMap) GetFlowWorkerStream(flowID int64) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	workerFlowID, exists := m.streamWorkerStream[flowID]
	return workerFlowID, exists
}

func (m *flowWorkerMap) SetFlowWorker(flowID int64, workerID string, workerFlowID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.streamWorker[flowID] = workerID
	m.streamWorkerStream[flowID] = workerFlowID
}

func (m *flowWorkerMap) RemoveFlow(flowID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.streamWorker, flowID)
	delete(m.streamWorkerStream, flowID)
}

func (m *flowWorkerMap) RemoveFlowIfMatches(flowID int64, workerFlowID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if currentWSID, exists := m.streamWorkerStream[flowID]; exists && currentWSID == workerFlowID {
		delete(m.streamWorker, flowID)
		delete(m.streamWorkerStream, flowID)
	}
}
