package performancemetrics

import (
	"context"

	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"golang.org/x/time/rate"
)

var explainLimiter = rate.NewLimiter(rate.Limit(commonutils.ExplainTPS), commonutils.ExplainTPS)

func WaitExplain(ctx context.Context) error {
	return explainLimiter.Wait(ctx)
}
