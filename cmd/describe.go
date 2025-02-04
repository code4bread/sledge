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

// describeCmd retrieves details about a Cloud SQL instance in pure JSON
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a Cloud SQL instance (outputs valid JSON only)",
	RunE:  runDescribe,
}

func init() {
	describeCmd.Flags().String("project", "", "GCP Project ID (required)")
	describeCmd.Flags().String("instance", "", "Name of the Cloud SQL instance (required)")

	viper.BindPFlag("describe.project", describeCmd.Flags().Lookup("project"))
	viper.BindPFlag("describe.instance", describeCmd.Flags().Lookup("instance"))
}

// runDescribe strictly prints JSON so that callers (e.g., a K8s operator) can parse it 
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

	// Retrieve instance details from GCP
	inst, err := sqlService.Instances.Get(projectID, instanceName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error describing instance %s: %v", instanceName, err)
	}

	// Marshal the entire struct to JSON
	data, marshalErr := json.MarshalIndent(inst, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal instance details: %v", marshalErr)
	}

	// Print JSON without additional text/log lines
	fmt.Println(string(data))
	return nil
}