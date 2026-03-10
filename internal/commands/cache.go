package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nicolasacchi/gumlet/internal/client"
	"github.com/spf13/cobra"
)

var (
	cacheURL       string
	cacheURLs      string
	cacheFile      string
	cacheStdin     bool
	cachePath      string
	cacheDryRun    bool
	cacheBatchSize int
)

type PurgeResult struct {
	URL    string `json:"url"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cachePurgeCmd)

	cachePurgeCmd.Flags().StringVar(&cacheURL, "url", "", "Single URL to purge")
	cachePurgeCmd.Flags().StringVar(&cacheURLs, "urls", "", "Comma-separated URLs to purge")
	cachePurgeCmd.Flags().StringVar(&cacheFile, "file", "", "File with URLs to purge (one per line)")
	cachePurgeCmd.Flags().BoolVar(&cacheStdin, "stdin", false, "Read URLs from stdin (one per line)")
	cachePurgeCmd.Flags().StringVar(&cachePath, "path", "", "Path pattern to purge (requires --subdomain)")
	cachePurgeCmd.Flags().BoolVar(&cacheDryRun, "dry-run", false, "Show URLs that would be purged without executing")
	cachePurgeCmd.Flags().IntVar(&cacheBatchSize, "batch-size", 50, "URLs per API call")
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Cache management operations",
}

var cachePurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge cached images",
	Long: `Purge one or more cached image URLs from Gumlet CDN.

Examples:
  gumlet cache purge --url https://mysub.gumlet.io/image.jpg
  gumlet cache purge --urls https://a.gumlet.io/1.jpg,https://a.gumlet.io/2.jpg
  gumlet cache purge --file urls.txt
  cat urls.txt | gumlet cache purge --stdin
  gumlet cache purge --path "/products/12345.jpg" -s mysub`,
	RunE: func(cmd *cobra.Command, args []string) error {
		subdomain := getSubdomain()

		urls, err := collectURLs(cacheURL, cacheURLs, cacheFile, cacheStdin, cachePath, subdomain)
		if err != nil {
			return err
		}

		if len(urls) == 0 {
			return fmt.Errorf("no URLs provided — use --url, --urls, --file, --stdin, or --path")
		}

		if cacheDryRun {
			if isJSONMode() {
				var results []PurgeResult
				for _, u := range urls {
					results = append(results, PurgeResult{URL: u, Status: "dry-run"})
				}
				data, _ := json.Marshal(results)
				return printData("cache.purge", data)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Would purge %d URL(s):\n", len(urls))
			for _, u := range urls {
				fmt.Fprintln(cmd.OutOrStdout(), "  "+u)
			}
			return nil
		}

		if subdomain == "" {
			// Try to extract subdomain from the first URL
			subdomain = extractSubdomain(urls[0])
			if subdomain == "" {
				return fmt.Errorf("--subdomain is required for cache purge (use --subdomain flag or 'gumlet config add' with subdomain)")
			}
		}

		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		results, err := batchPurge(context.Background(), c, subdomain, urls, cacheBatchSize)
		if err != nil {
			return err
		}

		data, _ := json.Marshal(results)
		if err := printData("cache.purge", data); err != nil {
			return err
		}

		// Count failures
		failures := 0
		for _, r := range results {
			if r.Status == "failed" {
				failures++
			}
		}
		if failures > 0 {
			return fmt.Errorf("%d of %d URLs failed to purge", failures, len(results))
		}
		return nil
	},
}

func collectURLs(url, urls, file string, stdin bool, path, subdomain string) ([]string, error) {
	seen := make(map[string]bool)
	var result []string

	add := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			return
		}
		seen[u] = true
		result = append(result, u)
	}

	if url != "" {
		add(url)
	}

	if urls != "" {
		for _, u := range strings.Split(urls, ",") {
			add(u)
		}
	}

	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("open file %s: %w", file, err)
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			add(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("read file %s: %w", file, err)
		}
	}

	if stdin {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			add(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
	}

	if path != "" {
		if subdomain == "" {
			return nil, fmt.Errorf("--subdomain is required when using --path")
		}
		fullURL := fmt.Sprintf("https://%s.gumlet.io%s", subdomain, path)
		add(fullURL)
	}

	return result, nil
}

func batchPurge(ctx context.Context, c *client.Client, subdomain string, urls []string, batchSize int) ([]PurgeResult, error) {
	var allResults []PurgeResult

	for i := 0; i < len(urls); i += batchSize {
		end := i + batchSize
		if end > len(urls) {
			end = len(urls)
		}
		batch := urls[i:end]

		body := map[string]any{
			"urls": batch,
		}

		_, err := c.Post(ctx, fmt.Sprintf("v1/purge/%s", subdomain), body)
		if err != nil {
			// Mark all URLs in this batch as failed
			for _, u := range batch {
				allResults = append(allResults, PurgeResult{
					URL:    u,
					Status: "failed",
					Error:  err.Error(),
				})
			}
			continue
		}

		for _, u := range batch {
			allResults = append(allResults, PurgeResult{
				URL:    u,
				Status: "purged",
			})
		}
	}

	return allResults, nil
}

func extractSubdomain(rawURL string) string {
	if !strings.Contains(rawURL, ".gumlet.io") {
		return ""
	}
	// Extract subdomain from https://subdomain.gumlet.io/...
	rawURL = strings.TrimPrefix(rawURL, "https://")
	rawURL = strings.TrimPrefix(rawURL, "http://")
	parts := strings.SplitN(rawURL, ".", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
