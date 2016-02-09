package monitor

import (
	"errors"
	"sync"
	"time"

	proto "github.com/micro/go-platform/monitor/proto"
	"golang.org/x/net/context"
)

type monitor struct {
	sync.RWMutex
	healthChecks map[string][]*proto.HealthCheck
	services     map[string]*proto.Service
}

var (
	DefaultMonitor   = newMonitor()
	ErrNotFound      = errors.New("not found")
	HealthCheckTopic = "micro.monitor.healthcheck"
	TickInterval     = time.Duration(time.Minute * 10)
)

func newMonitor() *monitor {
	return &monitor{
		healthChecks: make(map[string][]*proto.HealthCheck),
		services:     make(map[string]*proto.Service),
	}
}

func filter(hc []*proto.HealthCheck, status proto.HealthCheck_Status, limit, offset int) []*proto.HealthCheck {
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

func (m *monitor) reap() {
	m.Lock()
	defer m.Unlock()

	t := time.Now().Unix()

	services := make(map[string]*proto.Service)

	for id, hc := range m.healthChecks {
		var checks []*proto.HealthCheck
		for _, check := range hc {
			if t > (check.Timestamp+check.Interval) && t > (check.Timestamp+check.Ttl) {
				continue
			}
			checks = append(checks, check)

			// create new service list
			// TODO: maybe hold onto it so we have history
			if check.Service != nil && len(check.Service.Name) > 0 {
				if len(check.Service.Nodes) > 0 {
					services[check.Service.Nodes[0].Id] = check.Service
				}
			}
		}
		m.healthChecks[id] = checks
	}

	m.services = services
}

func (m *monitor) run() {
	t := time.NewTicker(TickInterval)

	for _ = range t.C {
		m.reap()
	}
}

func (m *monitor) HealthChecks(id string, status proto.HealthCheck_Status, limit, offset int) ([]*proto.HealthCheck, error) {
	m.RLock()
	defer m.RUnlock()

	if len(id) == 0 {
		var hcs []*proto.HealthCheck
		for _, hc := range m.healthChecks {
			hcs = append(hcs, hc...)
		}
		return filter(hcs, status, limit, offset), nil
	}

	hcs, ok := m.healthChecks[id]
	if !ok {
		return nil, ErrNotFound
	}
	return filter(hcs, status, limit, offset), nil
}

func (m *monitor) Services(s string) ([]*proto.Service, error) {
	m.RLock()
	defer m.RUnlock()

	toService := make(map[string][]*proto.Service)

	for _, service := range m.services {
		if len(s) > 0 && service.Name != s {
			continue
		}

		cp := &proto.Service{}

		*cp = *service

		sr, ok := toService[cp.Name]
		if !ok {
			toService[cp.Name] = []*proto.Service{cp}
			continue
		}

		// insert nodes into service version
		var seen bool
		for _, srv := range sr {
			if srv.Version == cp.Version {
				srv.Nodes = append(srv.Nodes, cp.Nodes...)
				seen = true
				break
			}
		}
		if !seen {
			toService[cp.Name] = append(toService[cp.Name], cp)
		}
	}

	var services []*proto.Service
	for _, service := range toService {
		services = append(services, service...)
	}
	return services, nil
}

func (m *monitor) ProcessHealthCheck(ctx context.Context, hc *proto.HealthCheck) error {
	m.Lock()
	defer m.Unlock()

	if hc.Service != nil && len(hc.Service.Name) > 0 {
		if len(hc.Service.Nodes) > 0 {
			m.services[hc.Service.Nodes[0].Id] = hc.Service
		}
	}

	hcs, ok := m.healthChecks[hc.Id]
	if !ok {
		m.healthChecks[hc.Id] = append(m.healthChecks[hc.Id], hc)
		return nil
	}

	for i, h := range hcs {
		if len(hc.Service.Nodes) > 0 && (h.Service.Nodes[0].Id == hc.Service.Nodes[0].Id) {
			hcs[i] = hc
			m.healthChecks[hc.Id] = hcs
			return nil
		}
	}

	hcs = append(hcs, hc)
	m.healthChecks[hc.Id] = hcs
	return nil
}

func (m *monitor) Run() {
	go m.run()
}
