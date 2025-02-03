package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"google.golang.org/api/sqladmin/v1"
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Cloud SQL instance",
	RunE:  runCreate,
}

func init() {
	// Flags for creating an instance
	CreateCmd.Flags().String("project", "", "GCP Project ID (required)")
	CreateCmd.Flags().String("instance", "", "Name of the new Cloud SQL instance (required)")
	CreateCmd.Flags().String("tier", "db-f1-micro", "Machine type tier for MySQL (e.g. db-g1-small, db-f1-micro)")
	CreateCmd.Flags().String("region", "us-central1", "Region for the instance")
	CreateCmd.Flags().String("dbVersion", "MYSQL_8_0", "Database version, e.g. MYSQL_5_7 or MYSQL_8_0")

	// Bind flags to viper
	viper.BindPFlag("create.project", CreateCmd.Flags().Lookup("project"))
	viper.BindPFlag("create.instance", CreateCmd.Flags().Lookup("instance"))
	viper.BindPFlag("create.tier", CreateCmd.Flags().Lookup("tier"))
	viper.BindPFlag("create.region", CreateCmd.Flags().Lookup("region"))
	viper.BindPFlag("create.dbVersion", CreateCmd.Flags().Lookup("dbVersion"))
}

func runCreate(cmd *cobra.Command, args []string) error {
	projectID := viper.GetString("create.project")
	instanceName := viper.GetString("create.instance")
	tier := viper.GetString("create.tier")
	region := viper.GetString("create.region")
	dbVersion := viper.GetString("create.dbVersion")

	if projectID == "" || instanceName == "" {
		return fmt.Errorf("project and instance flags are required")
	}

	ctx := context.Background()


	sqlService, err := sqladmin.NewService(ctx, option.WithScopes(sqladmin.CloudPlatformScope))
	if err != nil {
		return fmt.Errorf("failed to create sql admin service: %v", err)
	}

	instance := &sqladmin.DatabaseInstance{
		Name:            instanceName,
		Region:          region,
		DatabaseVersion: dbVersion,
		Settings: &sqladmin.Settings{
			Tier: tier,
		},
	}

	op, err := sqlService.Instances.Insert(projectID, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error creating instance: %v", err)
	}

	log.Printf("Creation initiated for instance %s. Operation: %s\n", instanceName, op.Name)
	return nil
}
