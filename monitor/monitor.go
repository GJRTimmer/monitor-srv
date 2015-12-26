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

func (m *monitor) HealthChecks(id string) ([]*proto.HealthCheck, error) {
	m.Lock()
	defer m.Unlock()

	hcs, ok := m.healthChecks[id]
	if !ok {
		return nil, ErrNotFound
	}
	return hcs, nil
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
