package commonutils

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	commonparams "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
)

var (
	typeCache   sync.Map // reflect.Type → []fieldDesc
	entityCache sync.Map // "host:port" → *integration.Entity
)

type fieldDesc struct {
	name string
	kind metric.SourceType
	idx  int
}

func describe(t reflect.Type) []fieldDesc {
	if v, ok := typeCache.Load(t); ok {
		return v.([]fieldDesc)
	}
	var list []fieldDesc
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("ingest_data") == "false" {
			continue
		}
		st := metric.GAUGE
		if f.Tag.Get("source_type") == "attribute" {
			st = metric.ATTRIBUTE
		}
		list = append(list, fieldDesc{
			name: f.Tag.Get("metric_name"), kind: st, idx: i,
		})
	}
	typeCache.Store(t, list)
	return list
}

func ProcessModel(model interface{}, ms *metric.Set) error {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return ErrInvalidModelType
	}
	for _, fd := range describe(val.Type()) {
		fv := val.Field(fd.idx)
		if fv.Kind() == reflect.Ptr && fv.IsNil() {
			continue
		}
		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}
		if err := ms.SetMetric(fd.name, fv.Interface(), fd.kind); err != nil {
			log.Debug("setMetric %s: %v", fd.name, err)
		}
	}
	return nil
}

func CreateEntity(pgInt *integration.Integration, cp *commonparams.CommonParameters) (*integration.Entity, error) {
	key := fmt.Sprintf("%s:%s", cp.Host, cp.Port)
	if v, ok := entityCache.Load(key); ok {
		return v.(*integration.Entity), nil
	}
	ent, err := pgInt.Entity(key, "pg-instance")
	if err != nil {
		return nil, err
	}
	entityCache.Store(key, ent)
	return ent, nil
}

func IngestMetric(list []interface{}, evt string, pgInt *integration.Integration, cp *commonparams.CommonParameters) error {
	ent, err := CreateEntity(pgInt, cp)
	if err != nil {
		return err
	}
	batch := 0
	for _, m := range list {
		if m == nil {
			continue
		}
		ms := ent.NewMetricSet(evt)
		if err := ProcessModel(m, ms); err != nil {
			log.Error("ProcessModel: %v", err)
		}
		batch++
		if batch >= PublishThreshold {
			if err := pgInt.Publish(); err != nil {
				return err
			}
			batch = 0
		}
	}
	if batch > 0 {
		return pgInt.Publish()
	}
	return nil
}
