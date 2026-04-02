package analytics

import "time"

type FlowStatusCount struct {
	Status string
	Count  int64
}

type ComponentCount struct {
	Component string
	Count     int64
}

type TimeSeriesPoint struct {
	Timestamp   time.Time
	InputEvents int64
	OutputEvents int64
	ErrorEvents int64
}

type Result struct {
	TotalFlows         int64
	FlowsByStatus      []FlowStatusCount
	TotalInputEvents     uint64
	TotalOutputEvents    uint64
	TotalProcessorErrors uint64
	ActiveWorkers        int64
	TotalEvents          int64
	ErrorEvents          int64
	EventsOverTime       []TimeSeriesPoint
	TopInputComponents   []ComponentCount
	TopOutputComponents  []ComponentCount
}

type Provider interface {
	GetAnalytics() (*Result, error)
}
