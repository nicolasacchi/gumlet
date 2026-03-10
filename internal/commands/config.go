package commands

import (
	"encoding/json"
	"fmt"

	"github.com/nicolasacchi/gumlet/internal/config"
	"github.com/spf13/cobra"
)

var (
	configAPIKeyFlag    string
	configSubdomainFlag string
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRemoveCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configUseCmd)
	configCmd.AddCommand(configCurrentCmd)

	configAddCmd.Flags().StringVar(&configAPIKeyFlag, "api-key", "", "API key for this project (required)")
	configAddCmd.Flags().StringVar(&configSubdomainFlag, "subdomain", "", "Default subdomain for this project")
	configAddCmd.MarkFlagRequired("api-key")
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add or update a named project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.AddProject(name, configAPIKeyFlag, configSubdomainFlag); err != nil {
			return err
		}
		if !quietFlag {
			fmt.Fprintf(cmd.OutOrStdout(), "Project %q added (key: %s)\n", name, config.MaskKey(configAPIKeyFlag))
		}
		return nil
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a named project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.RemoveProject(name); err != nil {
			return err
		}
		if !quietFlag {
			fmt.Fprintf(cmd.OutOrStdout(), "Project %q removed\n", name)
		}
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ListProjects()
		if err != nil {
			return fmt.Errorf("no config file found — run 'gumlet config add' first")
		}

		if cfg.Projects == nil || len(cfg.Projects) == 0 {
			if !quietFlag {
				fmt.Fprintln(cmd.OutOrStdout(), "No projects configured. Run 'gumlet config add <name> --api-key <key>' to add one.")
			}
			return nil
		}

		var rows []map[string]any
		for name, p := range cfg.Projects {
			isDefault := "no"
			if name == cfg.DefaultProject {
				isDefault = "yes"
			}
			rows = append(rows, map[string]any{
				"name":      name,
				"api_key":   config.MaskKey(p.APIKey),
				"subdomain": p.Subdomain,
				"default":   isDefault,
			})
		}

		data, _ := json.Marshal(rows)
		return printData("config.list", data)
	},
}

var configUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the default project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.SetDefaultProject(name); err != nil {
			return err
		}
		if !quietFlag {
			fmt.Fprintf(cmd.OutOrStdout(), "Default project set to %q\n", name)
		}
		return nil
	},
}

var configCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the current active project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ListProjects()
		if err != nil {
			return fmt.Errorf("no config file found")
		}
		if cfg.DefaultProject == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "No default project set")
			return nil
		}
		p := cfg.Projects[cfg.DefaultProject]
		if isJSONMode() {
			row := map[string]any{
				"name":      cfg.DefaultProject,
				"api_key":   config.MaskKey(p.APIKey),
				"subdomain": p.Subdomain,
			}
			data, _ := json.Marshal(row)
			return printData("", data)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Project:   %s\n", cfg.DefaultProject)
		fmt.Fprintf(cmd.OutOrStdout(), "API Key:   %s\n", config.MaskKey(p.APIKey))
		if p.Subdomain != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Subdomain: %s\n", p.Subdomain)
		}
		return nil
	},
}
