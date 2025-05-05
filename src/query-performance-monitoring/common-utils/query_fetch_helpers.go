package commonutils

import "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"

func FetchVersionSpecificSlowQueries(v uint64) (string, error) {
	switch {
	case v == PostgresVersion12:
		return queries.SlowQueriesForV12, nil
	case v >= PostgresVersion13:
		return queries.SlowQueriesForV13AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}

func FetchVersionSpecificBlockingQueries(v uint64) (string, error) {
	switch {
	case v == PostgresVersion12 || v == PostgresVersion13:
		return queries.BlockingQueriesForV12AndV13, nil
	case v >= PostgresVersion14:
		return queries.BlockingQueriesForV14AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}

func FetchVersionSpecificIndividualQueries(v uint64) (string, error) {
	switch {
	case v == PostgresVersion12:
		return queries.IndividualQuerySearchV12, nil
	case v > PostgresVersion12:
		return queries.IndividualQuerySearchV13AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}
