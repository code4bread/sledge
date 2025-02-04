package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"google.golang.org/api/sqladmin/v1"
)

// describeCmd retrieves details about a Cloud SQL instance
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a Cloud SQL instance",
	RunE:  runDescribe,
}

func init() {
	describeCmd.Flags().String("project", "", "GCP Project ID (required)")
	describeCmd.Flags().String("instance", "", "Name of the Cloud SQL instance (required)")

	viper.BindPFlag("describe.project", describeCmd.Flags().Lookup("project"))
	viper.BindPFlag("describe.instance", describeCmd.Flags().Lookup("instance"))
}

// runDescribe handles the logic for the 'describe' subcommand
func runDescribe(cmd *cobra.Command, args []string) error {
	projectID := viper.GetString("describe.project")
	instanceName := viper.GetString("describe.instance")

	if projectID == "" || instanceName == "" {
		return fmt.Errorf("both --project and --instance flags are required")
	}

	ctx := context.Background()
	sqlService, err := sqladmin.NewService(ctx, option.WithScopes(sqladmin.CloudPlatformScope))
	if err != nil {
		return fmt.Errorf("failed to create sql admin service: %v", err)
	}

	// Retrieve Cloud SQL instance details
	inst, err := sqlService.Instances.Get(projectID, instanceName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error describing instance %s: %v", instanceName, err)
	}

	// Print instance info in JSON or some user-friendly format
	data, err := json.MarshalIndent(inst, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal instance details: %v", err)
	}

	log.Printf("Cloud SQL Instance details for [%s] in project [%s]:\n%s\n", instanceName, projectID, string(data))
	return nil
}