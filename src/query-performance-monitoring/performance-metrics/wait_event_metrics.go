package performancemetrics

import (
	"context"
	"fmt"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	connpkg "github.com/newrelic/nri-postgresql/src/connection"
	commonparams "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/selfmetrics"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func PopulateWaitEventMetrics(ctx context.Context, conn *connpkg.PGSQLConnection, pgInt *integration.Integration, cp *commonparams.CommonParameters, exts map[string]bool) error {
	if ok, _ := validations.CheckWaitEventMetricsFetchEligibility(exts); !ok {
		return nil
	}
	if len(cp.Databases) == 0 {
		return nil
	}

	// Increment self-metrics counter
	selfmetrics.IncQueries()

	iface, err := getWaitEventMetrics(ctx, conn, cp)
	if err != nil {
		log.Error("wait-event fetch: %v", err)
		return err
	}
	if len(iface) == 0 {
		return nil
	}

	return commonutils.IngestMetric(iface, "PostgresWaitEvents", pgInt, cp)
}

func getWaitEventMetrics(ctx context.Context, conn *connpkg.PGSQLConnection, cp *commonparams.CommonParameters) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	query := fmt.Sprintf(queries.WaitEvents, cp.Databases, cp.QueryMonitoringCountThreshold)
	rows, err := conn.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []interface{}
	for rows.Next() {
		var m datamodels.WaitEventMetrics
		if err := rows.StructScan(&m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}
