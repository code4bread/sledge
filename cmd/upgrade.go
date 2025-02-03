package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"google.golang.org/api/sqladmin/v1"
)

var UpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade a Cloud SQL instance version or tier",
	RunE:  runUpgrade,
}

func init() {
	UpgradeCmd.Flags().String("project", "", "GCP Project ID (required)")
	UpgradeCmd.Flags().String("instance", "", "Name of the existing Cloud SQL instance (required)")
	UpgradeCmd.Flags().String("dbVersion", "", "New Database version, e.g. MYSQL_8_0")
	UpgradeCmd.Flags().String("tier", "", "New Machine type tier (optional)")

	viper.BindPFlag("upgrade.project", UpgradeCmd.Flags().Lookup("project"))
	viper.BindPFlag("upgrade.instance", UpgradeCmd.Flags().Lookup("instance"))
	viper.BindPFlag("upgrade.dbVersion", UpgradeCmd.Flags().Lookup("dbVersion"))
	viper.BindPFlag("upgrade.tier", UpgradeCmd.Flags().Lookup("tier"))
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	projectID := viper.GetString("upgrade.project")
	instanceName := viper.GetString("upgrade.instance")
	newVersion := viper.GetString("upgrade.dbVersion")
	newTier := viper.GetString("upgrade.tier")

	if projectID == "" || instanceName == "" {
		return fmt.Errorf("project and instance are required")
	}

	ctx := context.Background()
	sqlService, err := sqladmin.NewService(ctx, option.WithScopes(sqladmin.CloudPlatformScope))
	if err != nil {
		return fmt.Errorf("failed to create sql admin service: %v", err)
	}

	// Retrieve current instance
	currentInst, err := sqlService.Instances.Get(projectID, instanceName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("could not find instance %s: %v", instanceName, err)
	}

	// Update the version if provided
	if newVersion != "" {
		currentInst.DatabaseVersion = newVersion
	}

	// Update the tier if provided
	if newTier != "" {
		if currentInst.Settings == nil {
			currentInst.Settings = &sqladmin.Settings{}
		}
		currentInst.Settings.Tier = newTier
	}

	op, err := sqlService.Instances.Patch(projectID, instanceName, currentInst).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error updating instance: %v", err)
	}

	log.Printf("Upgrade initiated for instance %s. Operation: %s\n", instanceName, op.Name)
	return nil
}
