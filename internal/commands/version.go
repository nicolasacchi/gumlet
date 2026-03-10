package commands

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/nicolasacchi/gumlet/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		info := map[string]string{
			"version":    version,
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
		}

		if isJSONMode() {
			return output.PrintJSONValue(info)
		}

		data, _ := json.Marshal(info)
		fmt.Printf("gumlet %s (%s/%s, %s)\n", version, runtime.GOOS, runtime.GOARCH, runtime.Version())
		_ = data
		return nil
	},
}
