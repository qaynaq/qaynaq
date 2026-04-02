package analytics

import (
	"time"

	"github.com/qaynaq/qaynaq/internal/persistence"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type LocalProvider struct {
	db *gorm.DB
}

func NewLocalProvider(db *gorm.DB) Provider {
	return &LocalProvider{db: db}
}

func (p *LocalProvider) GetAnalytics() (*Result, error) {
	result := &Result{}

	g := new(errgroup.Group)
	g.Go(func() error { return p.loadFlowStats(result) })
	g.Go(func() error { return p.loadWorkerFlowMetrics(result) })
	g.Go(func() error { return p.loadActiveWorkers(result) })
	g.Go(func() error { return p.loadEventStats(result) })
	g.Go(func() error { return p.loadEventsOverTime(result) })
	g.Go(func() error { return p.loadTopComponents(result) })

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *LocalProvider) loadFlowStats(result *Result) error {
	type statusCount struct {
		Status string
		Count  int64
	}
	var counts []statusCount
	if err := p.db.Model(&persistence.Flow{}).
		Select("status, COUNT(*) as count").
		Where("is_current = true").
		Group("status").
		Find(&counts).Error; err != nil {
		return err
	}

	result.FlowsByStatus = make([]FlowStatusCount, len(counts))
	for i, c := range counts {
		result.FlowsByStatus[i] = FlowStatusCount{Status: c.Status, Count: c.Count}
		result.TotalFlows += c.Count
	}
	return nil
}

func (p *LocalProvider) loadWorkerFlowMetrics(result *Result) error {
	type metrics struct {
		TotalInput  uint64
		TotalOutput uint64
		TotalErrors uint64
	}
	var m metrics
	if err := p.db.Model(&persistence.WorkerFlow{}).
		Select("COALESCE(SUM(input_events), 0) as total_input, COALESCE(SUM(output_events), 0) as total_output, COALESCE(SUM(processor_errors), 0) as total_errors").
		Scan(&m).Error; err != nil {
		return err
	}
	result.TotalInputEvents = m.TotalInput
	result.TotalOutputEvents = m.TotalOutput
	result.TotalProcessorErrors = m.TotalErrors
	return nil
}

func (p *LocalProvider) loadActiveWorkers(result *Result) error {
	var count int64
	if err := p.db.Model(&persistence.Worker{}).
		Where("status = ?", persistence.WorkerStatusActive).
		Count(&count).Error; err != nil {
		return err
	}
	result.ActiveWorkers = count
	return nil
}

func (p *LocalProvider) loadEventStats(result *Result) error {
	var total int64
	if err := p.db.Model(&persistence.Event{}).Count(&total).Error; err != nil {
		return err
	}
	result.TotalEvents = total

	var errorCount int64
	if err := p.db.Model(&persistence.Event{}).
		Where("type = ?", persistence.EventTypeError).
		Count(&errorCount).Error; err != nil {
		return err
	}
	result.ErrorEvents = errorCount
	return nil
}

func (p *LocalProvider) loadEventsOverTime(result *Result) error {
	since := time.Now().AddDate(0, 0, -30)

	type dailyPoint struct {
		Day         string
		InputCount  int64
		OutputCount int64
		ErrorCount  int64
	}
	var points []dailyPoint

	if err := p.db.Model(&persistence.Event{}).
		Select(`DATE(created_at) as day,
			SUM(CASE WHEN section = 'input' THEN 1 ELSE 0 END) as input_count,
			SUM(CASE WHEN section = 'output' THEN 1 ELSE 0 END) as output_count,
			SUM(CASE WHEN type = 'ERROR' THEN 1 ELSE 0 END) as error_count`).
		Where("created_at >= ?", since).
		Group("DATE(created_at)").
		Order("day ASC").
		Find(&points).Error; err != nil {
		return err
	}

	result.EventsOverTime = make([]TimeSeriesPoint, len(points))
	for i, pt := range points {
		t, _ := time.Parse("2006-01-02", pt.Day)
		result.EventsOverTime[i] = TimeSeriesPoint{
			Timestamp:    t,
			InputEvents:  pt.InputCount,
			OutputEvents: pt.OutputCount,
			ErrorEvents:  pt.ErrorCount,
		}
	}
	return nil
}

func (p *LocalProvider) loadTopComponents(result *Result) error {
	type compCount struct {
		Component string
		Count     int64
	}

	var inputs []compCount
	if err := p.db.Model(&persistence.Flow{}).
		Select("input_component as component, COUNT(*) as count").
		Where("is_current = true").
		Group("input_component").
		Order("count DESC").
		Limit(10).
		Find(&inputs).Error; err != nil {
		return err
	}
	result.TopInputComponents = make([]ComponentCount, len(inputs))
	for i, c := range inputs {
		result.TopInputComponents[i] = ComponentCount{Component: c.Component, Count: c.Count}
	}

	var outputs []compCount
	if err := p.db.Model(&persistence.Flow{}).
		Select("output_component as component, COUNT(*) as count").
		Where("is_current = true").
		Group("output_component").
		Order("count DESC").
		Limit(10).
		Find(&outputs).Error; err != nil {
		return err
	}
	result.TopOutputComponents = make([]ComponentCount, len(outputs))
	for i, c := range outputs {
		result.TopOutputComponents[i] = ComponentCount{Component: c.Component, Count: c.Count}
	}

	return nil
}
