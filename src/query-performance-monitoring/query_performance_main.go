package queryperformancemonitoring

import (
	"context"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/collection"
	connpkg "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/metrics"

	commonparams "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/selfmetrics"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func QueryPerformanceMain(a args.ArgumentList, pgInt *integration.Integration, dbMap collection.DatabaseList) {
	if !a.EnableQueryMonitoring {
		log.Debug("query monitoring disabled by flag")
		return
	}
	if len(dbMap) == 0 {
		log.Debug("no databases found")
		return
	}

	connInfo := connpkg.DefaultConnectionInfo(&a)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := connInfo.NewConnection(connInfo.DatabaseName())
	if err != nil {
		log.Error("connection error: %v", err)
		return
	}
	defer db.Close()

	ver, err := metrics.CollectVersion(ctx, db)
	if err != nil {
		log.Error("version detect: %v", err)
		return
	}
	if !validations.CheckPostgresVersionSupportForQueryMonitoring(ver.Major) {
		log.Debug("Postgres %d not supported", ver.Major)
		return
	}

	cp := commonparams.SetCommonParameters(a, ver.Major, commonutils.GetDatabaseListInString(dbMap))
	populateQueryPerformance(ctx, db, pgInt, cp, connInfo)
}

func populateQueryPerformance(ctx context.Context, db *connpkg.PGSQLConnection, pgInt *integration.Integration, cp *commonparams.CommonParameters, info connpkg.Info) {
	exts, err := validations.FetchAllExtensions(db)
	if err != nil {
		log.Error("extension scan: %v", err)
		return
	}

	start := time.Now()
	slow := performancemetrics.PopulateSlowRunningMetrics(db, pgInt, cp, exts)
	selfmetrics.IncQueries()
	log.Debug("slow-running metrics in", time.Since(start))

	start = time.Now()
	_ = performancemetrics.PopulateWaitEventMetrics(ctx, db, pgInt, cp, exts)
	log.Debug("wait-event metrics in", time.Since(start))

	start = time.Now()
	performancemetrics.PopulateBlockingMetrics(ctx, db, pgInt, cp, exts)
	log.Debug("blocking metrics in", time.Since(start))

	start = time.Now()
	iq := performancemetrics.PopulateIndividualQueryMetrics(db, slow, pgInt, cp, exts)
	log.Debug("individual-query metrics in", time.Since(start))

	start = time.Now()
	performancemetrics.PopulateExecutionPlanMetrics(ctx, iq, pgInt, cp, info)
	log.Debug("execution-plan metrics in", time.Since(start))
}
