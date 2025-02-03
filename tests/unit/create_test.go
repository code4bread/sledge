package unit_test

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/code4bread/sledge/cmd"
)

func TestInit(t *testing.T) {
	// Call the init function to initialize flags and viper bindings
    createCmd :=cmd.CreateCmd

	// Check if the flags are correctly bound to viper
	assert.Equal(t, "", viper.GetString("create.project"))
	assert.Equal(t, "", viper.GetString("create.instance"))
	assert.Equal(t, "db-f1-micro", viper.GetString("create.tier"))
	assert.Equal(t, "us-central1", viper.GetString("create.region"))
	assert.Equal(t, "MYSQL_8_0", viper.GetString("create.dbVersion"))

	// Check if the flags are correctly set in the command
	projectFlag := createCmd.Flags().Lookup("project")
	assert.NotNil(t, projectFlag)
	assert.Equal(t, "GCP Project ID (required)", projectFlag.Usage)

	instanceFlag := createCmd.Flags().Lookup("instance")
	assert.NotNil(t, instanceFlag)
	assert.Equal(t, "Name of the new Cloud SQL instance (required)", instanceFlag.Usage)

	tierFlag := createCmd.Flags().Lookup("tier")
	assert.NotNil(t, tierFlag)
	assert.Equal(t, "Machine type tier for MySQL (e.g. db-g1-small, db-f1-micro)", tierFlag.Usage)

	regionFlag := createCmd.Flags().Lookup("region")
	assert.NotNil(t, regionFlag)
	assert.Equal(t, "Region for the instance", regionFlag.Usage)

	dbVersionFlag := createCmd.Flags().Lookup("dbVersion")
	assert.NotNil(t, dbVersionFlag)
	assert.Equal(t, "Database version, e.g. MYSQL_5_7 or MYSQL_8_0", dbVersionFlag.Usage)
}