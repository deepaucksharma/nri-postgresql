package inventory

import (
	"context"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestPopulateInventory(t *testing.T) {
	testIntegration, _ := integration.New("test", "0.1.0")
	testEntity, _ := testIntegration.Entity("test", "instance")

	testConnection, mock := connection.CreateMockSQL(t)
	ctx := context.Background()

	configRows := sqlmock.NewRows([]string{"name", "setting", "boot_val", "reset_val"}).
		AddRow("allow_system_table_mods", "off", "on", "test").
		AddRow("authentication_timeout", 1, 2, 3)

	// Use ExpectQuery instead of ExpectQueryContext since the older sqlmock doesn't have that method
	mock.ExpectQuery(configQuery).WillReturnRows(configRows)

	PopulateInventory(ctx, testEntity, testConnection)

	expected := inventory.Items{
		"allow_system_table_mods/setting": {
			"value": "off",
		},
		"allow_system_table_mods/boot_val": {
			"value": "on",
		},
		"allow_system_table_mods/reset_val": {
			"value": "test",
		},
		"authentication_timeout/setting": {
			"value": 1,
		},
		"authentication_timeout/boot_val": {
			"value": 2,
		},
		"authentication_timeout/reset_val": {
			"value": 3,
		},
	}

	assert.Equal(t, expected, testEntity.Inventory.Items())
}
