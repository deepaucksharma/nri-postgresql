package commonparameters

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
)

const (
	MaxQueryCountThreshold               = 30
	DefaultQueryMonitoringCountThreshold = 20
	DefaultQueryResponseTimeThreshold    = 500
)

type CommonParameters struct {
	Version                              uint64
	Databases                            string
	QueryMonitoringCountThreshold        int
	QueryMonitoringResponseTimeThreshold int
	Host                                 string
	Port                                 string
}

func SetCommonParameters(a args.ArgumentList, version uint64, dbs string) *CommonParameters {
	return &CommonParameters{
		Version:                              version,
		Databases:                            dbs,
		QueryMonitoringCountThreshold:        validateCount(a),
		QueryMonitoringResponseTimeThreshold: validateResponseTime(a),
		Host:                                 a.Hostname,
		Port:                                 a.Port,
	}
}

func validateCount(a args.ArgumentList) int {
	if a.QueryMonitoringCountThreshold < 0 {
		log.Warn("invalid count %d, using default %d", a.QueryMonitoringCountThreshold, DefaultQueryMonitoringCountThreshold)
		return DefaultQueryMonitoringCountThreshold
	}
	if a.QueryMonitoringCountThreshold > MaxQueryCountThreshold {
		log.Warn("count %d exceeds max %d", a.QueryMonitoringCountThreshold, MaxQueryCountThreshold)
		return MaxQueryCountThreshold
	}
	return a.QueryMonitoringCountThreshold
}

func validateResponseTime(a args.ArgumentList) int {
	if a.QueryMonitoringResponseTimeThreshold < 0 {
		log.Warn("invalid response time %d, using default %d", a.QueryMonitoringResponseTimeThreshold, DefaultQueryResponseTimeThreshold)
		return DefaultQueryResponseTimeThreshold
	}
	return a.QueryMonitoringResponseTimeThreshold
}
