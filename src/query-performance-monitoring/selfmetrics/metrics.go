package selfmetrics

import "sync/atomic"

var (
	qpQueriesScanned uint64
	qpExecPlans      uint64
	qpErrorsTotal    uint64
)

func IncQueries()      { atomic.AddUint64(&qpQueriesScanned, 1) }
func IncPlans()        { atomic.AddUint64(&qpExecPlans, 1) }
func IncErrors()       { atomic.AddUint64(&qpErrorsTotal, 1) }
func Snapshot() map[string]uint64 {
	return map[string]uint64{
		"qp_queries_scanned": atomic.LoadUint64(&qpQueriesScanned),
		"qp_exec_plans":      atomic.LoadUint64(&qpExecPlans),
		"qp_errors_total":    atomic.LoadUint64(&qpErrorsTotal),
	}
}
