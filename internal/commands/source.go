package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	sourceCreateName      string
	sourceCreateType      string
	sourceCreateNamespace string
	sourceCreateCDNType   string
	sourceCreateOriginURL string
	sourceCreateBucket    string
	sourceCreateRegion    string
	sourceCreateAccessKey string
	sourceCreateSecretKey string
)

func init() {
	rootCmd.AddCommand(sourceCmd)
	sourceCmd.AddCommand(sourceListCmd)
	sourceCmd.AddCommand(sourceDescribeCmd)
	sourceCmd.AddCommand(sourceCreateCmd)
	sourceCmd.AddCommand(sourceUpdateCmd)
	sourceCmd.AddCommand(sourceDeleteCmd)

	sourceCreateCmd.Flags().StringVar(&sourceCreateName, "name", "", "Source name (required)")
	sourceCreateCmd.Flags().StringVar(&sourceCreateType, "type", "", "Source type: s3, gcs, do_spaces, web_folder, custom (required)")
	sourceCreateCmd.Flags().StringVar(&sourceCreateNamespace, "namespace", "", "Subdomain namespace, becomes <namespace>.gumlet.io (required)")
	sourceCreateCmd.Flags().StringVar(&sourceCreateCDNType, "cdn-type", "gumlet", "CDN type (default: gumlet)")
	sourceCreateCmd.Flags().StringVar(&sourceCreateOriginURL, "origin-url", "", "Origin URL (for web_folder/custom)")
	sourceCreateCmd.Flags().StringVar(&sourceCreateBucket, "bucket", "", "Bucket name (for s3/gcs/do_spaces)")
	sourceCreateCmd.Flags().StringVar(&sourceCreateRegion, "region", "", "Region (for s3/do_spaces)")
	sourceCreateCmd.Flags().StringVar(&sourceCreateAccessKey, "access-key", "", "Access key (for s3/do_spaces)")
	sourceCreateCmd.Flags().StringVar(&sourceCreateSecretKey, "secret-key", "", "Secret key (for s3/do_spaces)")
	sourceCreateCmd.MarkFlagRequired("name")
	sourceCreateCmd.MarkFlagRequired("type")
	sourceCreateCmd.MarkFlagRequired("namespace")

	sourceUpdateCmd.Flags().StringVar(&sourceCreateName, "name", "", "New source name")
	sourceUpdateCmd.Flags().StringVar(&sourceCreateOriginURL, "origin-url", "", "New origin URL")
}

var sourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Manage image sources",
}

var sourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all image sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := c.Get(context.Background(), "v1/image/sources", nil)
		if err != nil {
			return err
		}
		// API returns {"all_sources": [...]} — unwrap to array
		var wrapper map[string]json.RawMessage
		if json.Unmarshal(data, &wrapper) == nil {
			if sources, ok := wrapper["all_sources"]; ok {
				data = sources
			}
		}
		return printData("source.list", data)
	},
}

var sourceDescribeCmd = &cobra.Command{
	Use:   "describe <source-id>",
	Short: "Show details for a specific source",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		data, err := c.Get(context.Background(), "v1/image/sources/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var sourceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new image source",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		// Map CLI type values to API type enum values
		apiType := sourceCreateType
		if apiType == "web_folder" {
			apiType = "webfolder"
		}

		body := map[string]any{
			"name":      sourceCreateName,
			"type":      apiType,
			"namespace": sourceCreateNamespace,
			"cdn_type":  sourceCreateCDNType,
		}

		// Build nested config object keyed by type
		switch apiType {
		case "webfolder", "custom":
			if sourceCreateOriginURL != "" {
				body[apiType] = map[string]any{"base_url": sourceCreateOriginURL}
			}
		case "s3":
			s3Config := map[string]any{}
			if sourceCreateBucket != "" {
				s3Config["bucket"] = sourceCreateBucket
			}
			if sourceCreateRegion != "" {
				s3Config["region"] = sourceCreateRegion
			}
			if sourceCreateAccessKey != "" {
				s3Config["access_key_id"] = sourceCreateAccessKey
			}
			if sourceCreateSecretKey != "" {
				s3Config["secret_access_key"] = sourceCreateSecretKey
			}
			body["s3"] = s3Config
		case "gcs":
			gcsConfig := map[string]any{}
			if sourceCreateBucket != "" {
				gcsConfig["bucket"] = sourceCreateBucket
			}
			body["gcs"] = gcsConfig
		case "do_spaces":
			doConfig := map[string]any{}
			if sourceCreateBucket != "" {
				doConfig["bucket"] = sourceCreateBucket
			}
			if sourceCreateRegion != "" {
				doConfig["region"] = sourceCreateRegion
			}
			if sourceCreateAccessKey != "" {
				doConfig["access_key_id"] = sourceCreateAccessKey
			}
			if sourceCreateSecretKey != "" {
				doConfig["secret_access_key"] = sourceCreateSecretKey
			}
			body["do_spaces"] = doConfig
		}

		data, err := c.Post(context.Background(), "v1/image/sources", body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var sourceUpdateCmd = &cobra.Command{
	Use:   "update <source-id>",
	Short: "Update an existing source",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		body := map[string]any{}
		if sourceCreateName != "" {
			body["name"] = sourceCreateName
		}
		if sourceCreateOriginURL != "" {
			body["origin_url"] = sourceCreateOriginURL
		}

		data, err := c.Post(context.Background(), "v1/image/sources/"+args[0], body)
		if err != nil {
			return err
		}
		return printData("", data)
	},
}

var sourceDeleteCmd = &cobra.Command{
	Use:   "delete <source-id>",
	Short: "Delete a source",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		if dryRun() {
			fmt.Fprintf(cmd.OutOrStdout(), "--dry-run: would delete source %s, no changes made\n", args[0])
			return nil
		}
		if err := requireConfirm(fmt.Sprintf("deleting source %s", args[0])); err != nil {
			return err
		}
		if err := c.Delete(context.Background(), "v1/image/sources/"+args[0]); err != nil {
			return err
		}
		if !quietFlag {
			fmt.Fprintf(cmd.OutOrStdout(), "Source %s deleted\n", args[0])
		}
		return nil
	},
}
