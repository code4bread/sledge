package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"google.golang.org/api/sqladmin/v1"
)

var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate a Cloud SQL instance from one region to another via backup & restore",
	RunE:  runMigrate,
}

func init() {
	MigrateCmd.Flags().String("sourceProject", "", "GCP Project ID of the source instance (required)")
	MigrateCmd.Flags().String("sourceInstance", "", "Source Cloud SQL instance name (required)")
	MigrateCmd.Flags().String("targetProject", "", "GCP Project ID for the new instance (defaults to sourceProject)")
	MigrateCmd.Flags().String("targetInstance", "", "Name of the new Cloud SQL instance in target region (required)")
	MigrateCmd.Flags().String("targetRegion", "", "Region where new instance should live (required)")
	MigrateCmd.Flags().String("backupDesc", "migration-backup", "Description for the on-demand backup")
	MigrateCmd.Flags().Duration("pollInterval", 5*time.Second, "Interval for polling operation status")
	MigrateCmd.Flags().Duration("pollTimeout", 10*time.Minute, "Timeout for polling operation completion")

	viper.BindPFlag("migrate.sourceProject", MigrateCmd.Flags().Lookup("sourceProject"))
	viper.BindPFlag("migrate.sourceInstance", MigrateCmd.Flags().Lookup("sourceInstance"))
	viper.BindPFlag("migrate.targetProject", MigrateCmd.Flags().Lookup("targetProject"))
	viper.BindPFlag("migrate.targetInstance", MigrateCmd.Flags().Lookup("targetInstance"))
	viper.BindPFlag("migrate.targetRegion", MigrateCmd.Flags().Lookup("targetRegion"))
	viper.BindPFlag("migrate.backupDesc", MigrateCmd.Flags().Lookup("backupDesc"))
	viper.BindPFlag("migrate.pollInterval", MigrateCmd.Flags().Lookup("pollInterval"))
	viper.BindPFlag("migrate.pollTimeout", MigrateCmd.Flags().Lookup("pollTimeout"))
}

func runMigrate(cmd *cobra.Command, args []string) error {
	sourceProject := viper.GetString("migrate.sourceProject")
	sourceInstance := viper.GetString("migrate.sourceInstance")
	targetProject := viper.GetString("migrate.targetProject")
	targetInstance := viper.GetString("migrate.targetInstance")
	targetRegion := viper.GetString("migrate.targetRegion")
	backupDesc := viper.GetString("migrate.backupDesc")
	pollInterval := viper.GetDuration("migrate.pollInterval")
	pollTimeout := viper.GetDuration("migrate.pollTimeout")

	if sourceProject == "" || sourceInstance == "" || targetInstance == "" || targetRegion == "" {
		return fmt.Errorf("sourceProject, sourceInstance, targetInstance, and targetRegion are required")
	}

	// If targetProject isn't provided, default it to sourceProject
	if targetProject == "" {
		targetProject = sourceProject
	}

	ctx := context.Background()
	sqlService, err := sqladmin.NewService(ctx, option.WithScopes(sqladmin.CloudPlatformScope))
	if err != nil {
		return fmt.Errorf("failed to create sql admin service: %v", err)
	}

	//
	// Step 1: Create On-Demand Backup of Source
	//
	log.Printf("[1/4] Creating on-demand backup for source instance %s...\n", sourceInstance)

	backupReq := &sqladmin.BackupRun{
		Description: backupDesc,
	}
	createBackupOp, err := sqlService.BackupRuns.Insert(sourceProject, sourceInstance, backupReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error creating on-demand backup: %v", err)
	}
	log.Printf("Backup operation started: %s\n", createBackupOp.Name)

	// Optionally poll for completion of backup
	backupErr := pollOperation(ctx, sqlService, sourceProject, createBackupOp.Name, pollInterval, pollTimeout)
	if backupErr != nil {
		return fmt.Errorf("backup operation failed or timed out: %v", backupErr)
	}
	log.Printf("[1/4] Backup complete.\n\n")

	// Retrieve the backupRunId we just created
	backupRunList, err := sqlService.BackupRuns.List(sourceProject, sourceInstance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to list backup runs: %v", err)
	}
	var latestBackup *sqladmin.BackupRun
	for _, br := range backupRunList.Items {
		if br.Description == backupDesc {
			latestBackup = br
			break
		}
	}
	if latestBackup == nil {
		return fmt.Errorf("could not find the backup run we created with description %s", backupDesc)
	}

	log.Printf("Using BackupRunId: %d\n\n", latestBackup.Id)

	//
	// Step 2: Fetch Source Instance Info
	//
	log.Printf("[2/4] Getting source instance info...\n")
	srcInst, err := sqlService.Instances.Get(sourceProject, sourceInstance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get source instance: %v", err)
	}
	log.Printf("[2/4] Source instance retrieved. DB Version: %s\n\n", srcInst.DatabaseVersion)

	//
	// Step 3: Create a new instance in target region (empty)
	//
	log.Printf("[3/4] Creating new instance %s in region %s...\n", targetInstance, targetRegion)

	newInst := &sqladmin.DatabaseInstance{
		Name:            targetInstance,
		Project:         targetProject,
		Region:          targetRegion,
		DatabaseVersion: srcInst.DatabaseVersion,
		Settings:        srcInst.Settings, // replicate same tier, flags, etc.
	}

	createInstOp, err := sqlService.Instances.Insert(targetProject, newInst).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error creating target instance: %v", err)
	}
	log.Printf("Creation operation started: %s\n", createInstOp.Name)

	// Poll creation
	createInstErr := pollOperation(ctx, sqlService, targetProject, createInstOp.Name, pollInterval, pollTimeout)
	if createInstErr != nil {
		return fmt.Errorf("instance creation failed or timed out: %v", createInstErr)
	}
	log.Printf("[3/4] Target instance created successfully.\n\n")

	//
	// Step 4: Restore Backup to the new instance
	//
	log.Printf("[4/4] Restoring backup ID %d from %s into %s...\n", latestBackup.Id, sourceInstance, targetInstance)
	restoreReq := &sqladmin.InstancesRestoreBackupRequest{
		RestoreBackupContext: &sqladmin.RestoreBackupContext{
			BackupRunId: latestBackup.Id,
			InstanceId:  sourceInstance,
			Project:     sourceProject, // project from which backup originated
		},
	}

	restoreOp, err := sqlService.Instances.RestoreBackup(targetProject, targetInstance, restoreReq).Context(ctx).Do()
	if err != nil {
		if strings.Contains(err.Error(), "not supported for cross region") {
			return fmt.Errorf("cross-region restore may not be supported for your DB version or region: %v", err)
		}
		return fmt.Errorf("failed to restore backup to new instance: %v", err)
	}
	log.Printf("Restore operation started: %s\n", restoreOp.Name)

	restoreErr := pollOperation(ctx, sqlService, targetProject, restoreOp.Name, pollInterval, pollTimeout)
	if restoreErr != nil {
		return fmt.Errorf("restore operation failed or timed out: %v", restoreErr)
	}

	log.Printf("[4/4] Migration complete. New instance: %s in region: %s\n", targetInstance, targetRegion)
	return nil
}

// pollOperation polls a long-running operation until completion or timeout
func pollOperation(ctx context.Context, sqlService *sqladmin.Service, projectID, operationName string,
	interval, timeout time.Duration) error {

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		op, err := sqlService.Operations.Get(projectID, operationName).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("failed to get operation %s: %v", operationName, err)
		}
		if op.Status == "DONE" {
			if op.Error != nil && len(op.Error.Errors) > 0 {
				return fmt.Errorf("operation %s finished with errors: %v", operationName, op.Error.Errors)
			}
			return nil
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("operation %s did not complete before timeout of %s", operationName, timeout)
}
 