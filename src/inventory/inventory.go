package inventory

import (
	"context"
	"time"
	
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/selfmetrics"
)

const (
	configQuery = `SELECT name, setting, boot_val, reset_val FROM pg_settings`
)

type configQueryRow struct {
	Name     string      `db:"name"`
	Setting  interface{} `db:"setting"`
	BootVal  interface{} `db:"boot_val"`
	ResetVal interface{} `db:"reset_val"`
}

// PopulateInventory collects all the configuration and populates the instance entity
func PopulateInventory(ctx context.Context, entity *integration.Entity, connection *connection.PGSQLConnection) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	selfmetrics.IncQueries()
	
	configRows := make([]*configQueryRow, 0)
	if err := connection.QueryContext(ctx, &configRows, configQuery); err != nil {
		log.Error("Failed to execute config query: %v", err)
	}

	for _, row := range configRows {
		logInventoryFailure(entity.SetInventoryItem(row.Name+"/setting", "value", row.Setting))
		logInventoryFailure(entity.SetInventoryItem(row.Name+"/boot_val", "value", row.BootVal))
		logInventoryFailure(entity.SetInventoryItem(row.Name+"/reset_val", "value", row.ResetVal))
	}
}

func logInventoryFailure(err error) {
	if err != nil {
		log.Error("Failed set inventory item: %v", err)
	}
}
