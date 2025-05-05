package performancemetrics

import (
	"context"
	"fmt"
	"time"

	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/selfmetrics"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

func PopulateBlockingMetrics(ctx context.Context, conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) {
	isEligible, enableCheckError := validations.CheckBlockingSessionMetricsFetchEligibility(enabledExtensions, cp.Version)
	if enableCheckError != nil {
		log.Error("Error executing query: %v in PopulateBlockingMetrics", enableCheckError)
		return
	}
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return
	}
	blockingQueriesMetricsList, blockQueryFetchErr := getBlockingMetrics(ctx, conn, cp)
	if blockQueryFetchErr != nil {
		log.Error("Error fetching Blocking queries: %v", blockQueryFetchErr)
		return
	}
	if len(blockingQueriesMetricsList) == 0 {
		log.Debug("No Blocking queries found.")
		return
	}
	err := commonutils.IngestMetric(blockingQueriesMetricsList, "PostgresBlockingSessions", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting Blocking queries: %v", err)
		return
	}
	
	// Increment self-metrics counter
	selfmetrics.IncQueries()
}

func getBlockingMetrics(ctx context.Context, conn *performancedbconnection.PGSQLConnection, cp *commonparameters.CommonParameters) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	var blockingQueriesMetricsList []interface{}
	versionSpecificBlockingQuery, err := commonutils.FetchVersionSpecificBlockingQueries(cp.Version)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return nil, err
	}
	var query = fmt.Sprintf(versionSpecificBlockingQuery, cp.Databases, cp.QueryMonitoringCountThreshold)
	rows, err := conn.QueryxContext(ctx, query)
	if err != nil {
		log.Error("Failed to execute query: %v", err)
		return nil, commonutils.ErrUnExpectedError
	}
	defer rows.Close()
	for rows.Next() {
		var blockingQueryMetric datamodels.BlockingSessionMetrics
		if scanError := rows.StructScan(&blockingQueryMetric); scanError != nil {
			return nil, scanError
		}
		// For PostgreSQL versions 13 and 12, anonymization of queries does not occur for blocking sessions, so it's necessary to explicitly anonymize them.
		if cp.Version == commonutils.PostgresVersion13 || cp.Version == commonutils.PostgresVersion12 {
			*blockingQueryMetric.BlockedQuery = commonutils.AnonymizeQueryText(*blockingQueryMetric.BlockedQuery)
			*blockingQueryMetric.BlockingQuery = commonutils.AnonymizeQueryText(*blockingQueryMetric.BlockingQuery)
		}
		blockingQueriesMetricsList = append(blockingQueriesMetricsList, blockingQueryMetric)
	}

	return blockingQueriesMetricsList, nil
}
