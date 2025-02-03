package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"google.golang.org/api/sqladmin/v1"
)

// RestoreCmd will restore a backup from an existing instance to a new or existing instance
var RestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a backup run into a Cloud SQL instance (possibly new instance in a different region)",
	RunE:  runRestore,
}

func init() {
	RestoreCmd.Flags().String("project", "", "GCP Project ID (required)")
	RestoreCmd.Flags().String("targetInstance", "", "Name of the instance to restore into (required)")
	RestoreCmd.Flags().Int64("backupRunId", 0, "BackupRun ID to restore from (required)")
	RestoreCmd.Flags().String("sourceInstance", "", "Name of the source instance from which backup was taken (required)")

	// Optionally, if you want to handle region changes in code,
	// you might add a --region for the new instance, but
	// the Cloud SQL Admin API might not let you do direct cross-region restore.

	viper.BindPFlag("restore.project", RestoreCmd.Flags().Lookup("project"))
	viper.BindPFlag("restore.targetInstance", RestoreCmd.Flags().Lookup("targetInstance"))
	viper.BindPFlag("restore.backupRunId", RestoreCmd.Flags().Lookup("backupRunId"))
	viper.BindPFlag("restore.sourceInstance", RestoreCmd.Flags().Lookup("sourceInstance"))
}

// runRestore calls Instances.RestoreBackup to restore from a specific backup run
func runRestore(cmd *cobra.Command, args []string) error {
	projectID := viper.GetString("restore.project")
	targetInstance := viper.GetString("restore.targetInstance")
	sourceInstance := viper.GetString("restore.sourceInstance")
	backupRunID := viper.GetInt64("restore.backupRunId")

	if projectID == "" || targetInstance == "" || backupRunID == 0 || sourceInstance == "" {
		return fmt.Errorf("project, targetInstance, sourceInstance, and backupRunId are required")
	}

	ctx := context.Background()
	sqlService, err := sqladmin.NewService(ctx, option.WithScopes(sqladmin.CloudPlatformScope))
	if err != nil {
		return fmt.Errorf("failed to create SQL Admin service: %v", err)
	}

	// "Migrate" scenario:
	// If you're trying to do cross-region, you might have to create the target instance with
	// the desired region first, then do restore.
	// This code assumes the targetInstance already exists or has the same region as source.

	req := &sqladmin.InstancesRestoreBackupRequest{
		RestoreBackupContext: &sqladmin.RestoreBackupContext{
			BackupRunId: backupRunID,
			InstanceId:  sourceInstance,
			Project:     projectID,
		},
	}

	op, err := sqlService.Instances.RestoreBackup(projectID, targetInstance, req).Context(ctx).Do()
	if err != nil {
		if strings.Contains(err.Error(), "not supported for cross region") {
			return fmt.Errorf("cross-region restore may not be supported for your DB version or region: %v", err)
		}
		return fmt.Errorf("error restoring backup to instance %s: %v", targetInstance, err)
	}

	log.Printf("Restore initiated for target instance %s from backup ID %d. Operation: %s\n",
		targetInstance, backupRunID, op.Name)
	return nil
}
