package handler

import (
	"github.com/micro/go-micro/errors"
	"github.com/micro/monitor-srv/monitor"
	proto "github.com/micro/monitor-srv/proto/monitor"
	"golang.org/x/net/context"
)

type Monitor struct{}

func (m *Monitor) HealthChecks(ctx context.Context, req *proto.HealthChecksRequest, rsp *proto.HealthChecksResponse) error {
	if req.Limit == 0 {
		req.Limit = 10
	}
	hcs, err := monitor.DefaultMonitor.HealthChecks(req.Id, req.Status, int(req.Limit), int(req.Offset))
	if err != nil && err == monitor.ErrNotFound {
		return errors.NotFound("go.micro.srv.monitor.Monitor.HealthCheck", err.Error())
	} else if err != nil {
		return errors.InternalServerError("go.micro.srv.monitor.Monitor.HealthCheck", err.Error())
	}

	rsp.Healthchecks = hcs
	return nil
}

func (m *Monitor) Services(ctx context.Context, req *proto.ServicesRequest, rsp *proto.ServicesResponse) error {
	services, err := monitor.DefaultMonitor.Services(req.Service)
	if err != nil {
		return errors.InternalServerError("go.micro.srv.monitor.Monitor.Services", err.Error())
	}
	rsp.Services = services
	return nil
}
