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
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func PopulateSlowRunningMetrics(conn *connpkg.PGSQLConnection, pgInt *integration.Integration, cp *commonparams.CommonParameters, exts map[string]bool) []datamodels.SlowRunningQueryMetrics {
	if ok, _ := validations.CheckSlowQueryMetricsFetchEligibility(exts); !ok {
		return nil
	}
	if len(cp.Databases) == 0 {
		return nil
	}

	list, iface, err := getSlowRunningMetrics(conn, cp)
	if err != nil {
		log.Error("slow query fetch: %v", err)
		return nil
	}
	if len(list) == 0 {
		return nil
	}

	if err := commonutils.IngestMetric(iface, "PostgresSlowQueries", pgInt, cp); err != nil {
		log.Error("ingest slow queries: %v", err)
	}
	return list
}

func getSlowRunningMetrics(conn *connpkg.PGSQLConnection, cp *commonparams.CommonParameters) ([]datamodels.SlowRunningQueryMetrics, []interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tpl, err := commonutils.FetchVersionSpecificSlowQueries(cp.Version)
	if err != nil {
		return nil, nil, err
	}

	query := fmt.Sprintf(tpl, cp.Databases, cp.QueryMonitoringCountThreshold)
	rows, err := conn.QueryxContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var list []datamodels.SlowRunningQueryMetrics
	var iface []interface{}

	for rows.Next() {
		var m datamodels.SlowRunningQueryMetrics
		if err := rows.StructScan(&m); err != nil {
			return nil, nil, err
		}
		list = append(list, m)
		iface = append(iface, m)
	}
	return list, iface, nil
}
