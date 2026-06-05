package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/nicolasacchi/gumlet/internal/client"
	"github.com/spf13/cobra"
)

var (
	transformWidth    int
	transformHeight   int
	transformFormat   string
	transformQuality  int
	transformCrop     string
	transformCompress bool
	transformBlur     int
	transformSharpen  bool
	transformMode     string
)

type TransformOptions struct {
	Width    int
	Height   int
	Format   string
	Quality  int
	Crop     string
	Compress bool
	Blur     int
	Sharpen  bool
	Mode     string
}

func init() {
	rootCmd.AddCommand(transformCmd)
	transformCmd.AddCommand(transformURLCmd)
	transformCmd.AddCommand(transformInspectCmd)

	transformURLCmd.Flags().IntVarP(&transformWidth, "width", "w", 0, "Target width")
	transformURLCmd.Flags().IntVar(&transformHeight, "height", 0, "Target height")
	transformURLCmd.Flags().StringVarP(&transformFormat, "format", "f", "", "Output format (webp, avif, auto, jpg, png, jxl)")
	transformURLCmd.Flags().IntVar(&transformQuality, "quality", 0, "Quality (1-100)")
	transformURLCmd.Flags().StringVar(&transformCrop, "crop", "", "Crop mode (smart, center, north, south, east, west)")
	transformURLCmd.Flags().BoolVar(&transformCompress, "compress", false, "Enable lossy compression")
	transformURLCmd.Flags().IntVar(&transformBlur, "blur", 0, "Blur radius")
	transformURLCmd.Flags().BoolVar(&transformSharpen, "sharpen", false, "Enable sharpening")
	transformURLCmd.Flags().StringVar(&transformMode, "mode", "", "Resize mode (fit, fill, stretch, pad)")
}

var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Image transform utilities",
}

var transformURLCmd = &cobra.Command{
	Use:   "url <base-url>",
	Short: "Generate a Gumlet transform URL",
	Long: `Build a Gumlet transform URL by appending query parameters.
This is a pure local command — no API call is made.

Examples:
  gumlet transform url https://mysub.gumlet.io/img.jpg --width 400 --format webp
  gumlet transform url https://mysub.gumlet.io/img.jpg --width 800 --quality 80 --crop smart`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := TransformOptions{
			Width:    transformWidth,
			Height:   transformHeight,
			Format:   transformFormat,
			Quality:  transformQuality,
			Crop:     transformCrop,
			Compress: transformCompress,
			Blur:     transformBlur,
			Sharpen:  transformSharpen,
			Mode:     transformMode,
		}

		result, err := buildTransformURL(args[0], opts)
		if err != nil {
			return err
		}

		if isJSONMode() {
			data := map[string]string{
				"original":    args[0],
				"transformed": result,
			}
			b, _ := json.Marshal(data)
			return printData("", b)
		}

		fmt.Fprintln(cmd.OutOrStdout(), result)
		return nil
	},
}

var transformInspectCmd = &cobra.Command{
	Use:   "inspect <url>",
	Short: "Inspect a Gumlet URL (cache status, format, size)",
	Long: `Fetch a Gumlet URL and report cache status, format served, size, and response time.

Example:
  gumlet transform inspect https://mysub.gumlet.io/img.jpg?w=400`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New("", "", verboseFlag)

		start := time.Now()
		resp, err := c.Head(args[0])
		if err != nil {
			return fmt.Errorf("fetch %s: %w", args[0], err)
		}
		elapsed := time.Since(start)

		// Read and discard body to get accurate content-length
		if resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		cacheStatus := resp.Header.Get("X-Gumlet-Cache")
		if cacheStatus == "" {
			cacheStatus = resp.Header.Get("X-Cache")
		}
		if cacheStatus == "" {
			cacheStatus = "unknown"
		}

		contentType := resp.Header.Get("Content-Type")
		contentLength := resp.Header.Get("Content-Length")
		sizeStr := contentLength
		if cl, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			sizeStr = humanSize(cl)
		}

		if isJSONMode() {
			data := []map[string]any{
				{"header": "Status", "value": fmt.Sprintf("%d", resp.StatusCode)},
				{"header": "Cache", "value": cacheStatus},
				{"header": "Format", "value": contentType},
				{"header": "Size", "value": sizeStr},
				{"header": "Response", "value": fmt.Sprintf("%dms", elapsed.Milliseconds())},
			}
			b, _ := json.Marshal(data)
			return printData("transform.inspect", b)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Status:    %d\n", resp.StatusCode)
		fmt.Fprintf(cmd.OutOrStdout(), "Cache:     %s\n", cacheStatus)
		fmt.Fprintf(cmd.OutOrStdout(), "Format:    %s\n", contentType)
		fmt.Fprintf(cmd.OutOrStdout(), "Size:      %s\n", sizeStr)
		fmt.Fprintf(cmd.OutOrStdout(), "Response:  %dms\n", elapsed.Milliseconds())

		// Print additional Gumlet-specific headers if present
		for _, h := range []string{"X-Gumlet-Process-Time", "X-Gumlet-Original-Size", "X-Gumlet-Optimized-Size"} {
			if v := resp.Header.Get(h); v != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "%-10s %s\n", h+":", v)
			}
		}

		return nil
	},
}

func buildTransformURL(base string, opts TransformOptions) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	q := u.Query()

	if opts.Width > 0 {
		q.Set("w", strconv.Itoa(opts.Width))
	}
	if opts.Height > 0 {
		q.Set("h", strconv.Itoa(opts.Height))
	}
	if opts.Format != "" {
		q.Set("format", opts.Format)
	}
	if opts.Quality > 0 {
		q.Set("q", strconv.Itoa(opts.Quality))
	}
	if opts.Crop != "" {
		q.Set("crop", opts.Crop)
	}
	if opts.Compress {
		q.Set("compress", "true")
	}
	if opts.Blur > 0 {
		q.Set("blur", strconv.Itoa(opts.Blur))
	}
	if opts.Sharpen {
		q.Set("sharpen", "true")
	}
	if opts.Mode != "" {
		q.Set("mode", opts.Mode)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func humanSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	kb := float64(bytes) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.1f KB", kb)
	}
	mb := kb / 1024
	return fmt.Sprintf("%.1f MB", mb)
}
