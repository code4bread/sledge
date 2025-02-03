package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"google.golang.org/api/sqladmin/v1"
)

// BackupCmd triggers an on-demand backup for an existing Cloud SQL instance.
var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create an on-demand backup for a Cloud SQL instance",
	RunE:  runBackup,
}

func init() {
	BackupCmd.Flags().String("project", "", "GCP Project ID (required)")
	BackupCmd.Flags().String("instance", "", "Name of the Cloud SQL instance (required)")
	BackupCmd.Flags().String("description", "on-demand-backup", "Description for this backup")

	viper.BindPFlag("backup.project", BackupCmd.Flags().Lookup("project"))
	viper.BindPFlag("backup.instance", BackupCmd.Flags().Lookup("instance"))
	viper.BindPFlag("backup.description", BackupCmd.Flags().Lookup("description"))

	// Register BackupCmd with the root command in root.go
}

func runBackup(cmd *cobra.Command, args []string) error {
	projectID := viper.GetString("backup.project")
	instanceName := viper.GetString("backup.instance")
	backupDescription := viper.GetString("backup.description")

	if projectID == "" || instanceName == "" {
		return fmt.Errorf("both --project and --instance are required")
	}

	ctx := context.Background()
	sqlService, err := sqladmin.NewService(ctx, option.WithScopes(sqladmin.CloudPlatformScope))
	if err != nil {
		return fmt.Errorf("failed to create SQL Admin service: %v", err)
	}

	backupRun := &sqladmin.BackupRun{
		Description: backupDescription,
	}
	op, err := sqlService.BackupRuns.Insert(projectID, instanceName, backupRun).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error creating backup for instance %s: %v", instanceName, err)
	}

	log.Printf("Backup initiated for instance %s. Operation: %s\n", instanceName, op.Name)
	return nil
}
