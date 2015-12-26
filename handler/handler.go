package handler

import (
	"github.com/micro/go-micro/errors"
	"github.com/micro/monitoring-srv/monitor"
	proto "github.com/micro/monitoring-srv/proto/monitor"
	"golang.org/x/net/context"
)

type Monitor struct{}

func (m *Monitor) HealthChecks(ctx context.Context, req *proto.HealthChecksRequest, rsp *proto.HealthChecksResponse) error {
	hcs, err := monitor.DefaultMonitor.HealthChecks(req.Id, req.Status, int(req.Limit), int(req.Offset))
	if err != nil && err == monitor.ErrNotFound {
		return errors.NotFound("go.micro.srv.monitoring.Monitor.HealthCheck", err.Error())
	} else if err != nil {
		return errors.InternalServerError("go.micro.srv.monitoring.Monitor.HealthCheck", err.Error())
	}

	rsp.HealthChecks = hcs
	return nil
}
