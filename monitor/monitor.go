package monitor

import (
	"errors"
	"sync"

	proto "github.com/micro/go-platform/monitor/proto"
	"golang.org/x/net/context"
)

type monitor struct {
	sync.RWMutex
	healthChecks map[string][]*proto.HealthCheck
}

var (
	ErrNotFound      = errors.New("not found")
	DefaultMonitor   = newMonitor()
	HealthCheckTopic = "micro.monitor.healthcheck"
)

func newMonitor() *monitor {
	return &monitor{
		healthChecks: make(map[string][]*proto.HealthCheck),
	}
}

func (m *monitor) filter(hc []*proto.HealthCheck, status proto.HealthCheck_Status, limit, offset int) []*proto.HealthCheck {
	if len(hc) < offset {
		return []*proto.HealthCheck{}
	}

	if (limit + offset) > len(hc) {
		limit = len(hc) - offset
	}

	var hcs []*proto.HealthCheck
	for i := 0; i < limit; i++ {
		if status == proto.HealthCheck_UNKNOWN {
			hcs = append(hcs, hc[offset])
		} else if hc[offset].Status == status {
			hcs = append(hcs, hc[offset])
		}
		offset++
	}
	return hcs
}

func (m *monitor) HealthChecks(id string, status proto.HealthCheck_Status, limit, offset int) ([]*proto.HealthCheck, error) {
	m.Lock()
	defer m.Unlock()

	if len(id) == 0 {
		var hcs []*proto.HealthCheck
		for _, hc := range m.healthChecks {
			hcs = append(hcs, hc...)
		}
		return m.filter(hcs, status, limit, offset), nil
	}

	hcs, ok := m.healthChecks[id]
	if !ok {
		return nil, ErrNotFound
	}
	return m.filter(hcs, status, limit, offset), nil
}

func (m *monitor) ProcessHealthCheck(ctx context.Context, hc *proto.HealthCheck) error {
	m.Lock()
	defer m.Unlock()

	hcs, ok := m.healthChecks[hc.Id]
	if !ok {
		m.healthChecks[hc.Id] = append(m.healthChecks[hc.Id], hc)
		return nil
	}

	for i, h := range hcs {
		if h.Service.Id == hc.Service.Id {
			hcs[i] = hc
			m.healthChecks[hc.Id] = hcs
			return nil
		}
	}

	hcs = append(hcs, hc)
	m.healthChecks[hc.Id] = hcs
	return nil
}
