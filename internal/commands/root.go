package commands

import (
	"github.com/nicolasacchi/gumlet/internal/client"
	"github.com/nicolasacchi/gumlet/internal/config"
	"github.com/nicolasacchi/gumlet/internal/output"
	"github.com/spf13/cobra"
)

var (
	version       = "dev"
	apiKeyFlag    string
	projectFlag   string
	subdomainFlag string
	jsonFlag      bool
	jqFlag        string
	verboseFlag   bool
	quietFlag     bool
)

var rootCmd = &cobra.Command{
	Use:   "gumlet",
	Short: "Gumlet CLI — image CDN management from the command line",
	Long: `gumlet is a CLI for the Gumlet Image CDN API. Manage sources,
purge caches, query analytics, and build transform URLs.

Usage examples:
  gumlet source list
  gumlet cache purge --url https://mysub.gumlet.io/image.jpg
  gumlet analytics summary --source-id abc123
  gumlet transform url https://mysub.gumlet.io/image.jpg --width 300 --format webp`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

func Execute() error {
	return rootCmd.Execute()
}

func getClient(cmd *cobra.Command) (*client.Client, error) {
	apiKey, err := config.LoadAPIKey(apiKeyFlag, projectFlag)
	if err != nil {
		return nil, err
	}
	return client.New(apiKey, "", verboseFlag), nil
}

func getSubdomain() string {
	return config.LoadSubdomain(subdomainFlag, projectFlag)
}

func isJSONMode() bool {
	return output.IsJSON(jsonFlag, jqFlag)
}

func printData(command string, data []byte) error {
	return output.PrintData(command, data, isJSONMode(), jqFlag)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "Gumlet API key (overrides GUMLET_API_KEY env var)")
	rootCmd.PersistentFlags().StringVar(&projectFlag, "project", "", "Use a named project from config file")
	rootCmd.PersistentFlags().StringVarP(&subdomainFlag, "subdomain", "s", "", "Gumlet subdomain (overrides config)")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Force JSON output (auto-enabled when stdout is not a TTY)")
	rootCmd.PersistentFlags().StringVar(&jqFlag, "jq", "", "Apply gjson path filter to JSON output")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Print request/response details to stderr")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Suppress non-error output")
}
