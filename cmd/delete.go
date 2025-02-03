package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"google.golang.org/api/sqladmin/v1"
)

// DeleteCmd removes an existing Cloud SQL instance
var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a Cloud SQL instance",
	RunE:  runDelete,
}

func init() {
	DeleteCmd.Flags().String("project", "", "GCP Project ID (required)")
	DeleteCmd.Flags().String("instance", "", "Name of the Cloud SQL instance to delete (required)")

	viper.BindPFlag("delete.project", DeleteCmd.Flags().Lookup("project"))
	viper.BindPFlag("delete.instance", DeleteCmd.Flags().Lookup("instance"))
}

func runDelete(cmd *cobra.Command, args []string) error {
	projectID := viper.GetString("delete.project")
	instanceName := viper.GetString("delete.instance")

	if projectID == "" || instanceName == "" {
		return fmt.Errorf("both --project and --instance flags are required")
	}

	// Create a context and the SQL Admin service
	ctx := context.Background()
	sqlService, err := sqladmin.NewService(ctx, option.WithScopes(sqladmin.CloudPlatformScope))
	if err != nil {
		return fmt.Errorf("failed to create sql admin service: %v", err)
	}

	// Attempt to delete the Cloud SQL instance
	op, err := sqlService.Instances.Delete(projectID, instanceName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error deleting instance %s: %v", instanceName, err)
	}

	log.Printf("Deletion initiated for instance %s. Operation: %s\n", instanceName, op.Name)
	return nil
}
