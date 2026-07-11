package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"jgo/internal/config"
	"jgo/internal/mirror"
	"jgo/internal/provider"

	"github.com/spf13/cobra"
)

var sourcetestProxy string

var sourcetestCmd = &cobra.Command{
	Use:   "sourcetest",
	Short: "Test connectivity to all JDK sources",
	Long: "Test connectivity to all JDK sources (Dragonwell, Corretto, Azul, Adoptium).\n" +
		"Sends the same requests as 'jgo list available' but outputs each request URL\n" +
		"and its response time instead of the JDK list.\n\n" +
		"  jgo sourcetest                       # test using configured proxy\n" +
		"  jgo sourcetest --proxy http://...    # test with a temporary proxy\n" +
		"  jgo sourcetest --proxy none          # test without proxy",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		proxy := cfg.Proxy
		if cmd.Flags().Changed("proxy") {
			if strings.ToLower(sourcetestProxy) == "none" {
				proxy = ""
			} else {
				proxy = sourcetestProxy
			}
		}

		osName := provider.MapOS()
		arch := provider.MapArch()
		if proxy != "" {
			fmt.Fprintf(os.Stderr, "Using proxy: %s\n", proxy)
		}
		fmt.Fprintf(os.Stderr, "Testing sources for %s/%s...\n\n", osName, arch)

		mirrors := mirror.Resolve(cfg.Mirrors)
		for source, mb := range mirrors.Backends {
			if m, ok := mb.(mirror.Mirror); ok {
				fmt.Fprintf(os.Stderr, "  %s source: %s (%s)\n", source, m.ID(), m.DisplayName())
			}
		}

		spinner := StartSpinner("Testing")
		wallStart := time.Now()
		results := provider.TestAllSources(osName, arch, proxy, mirrors)
		wallEnd := time.Now()
		spinner.Stop()

		var totalReqs int
		var sumDur time.Duration
		var failedReqs int

		for _, r := range results {
			if r.Error != nil {
				fmt.Printf("  %s: ERROR - %v\n", r.Source, r.Error)
				continue
			}
			if len(r.Records) == 0 {
				fmt.Printf("  %s: no requests\n", r.Source)
				continue
			}

			var srcMax time.Duration
			fmt.Printf("  %s (%d requests):\n", r.Source, len(r.Records))
			for _, rec := range r.Records {
				totalReqs++
				sumDur += rec.Duration
				if rec.Duration > srcMax {
					srcMax = rec.Duration
				}
				status := rec.Status
				if status == "OK" {
					status = "\033[32mOK\033[0m"
				} else {
					failedReqs++
					status = "\033[31m" + status + "\033[0m"
				}
				fmt.Printf("    %s  %8s  %s\n", rec.Duration.Round(time.Millisecond), status, shortenURL(rec.URL))
			}
			fmt.Printf("    -> slowest: %s\n\n", srcMax.Round(time.Millisecond))
		}

		wall := wallEnd.Sub(wallStart)
		fmt.Printf("  Total: %d requests, %d OK, %d failed\n", totalReqs, totalReqs-failedReqs, failedReqs)
		fmt.Printf("  Wall time: %s (concurrent), sum of requests: %s\n", wall.Round(time.Millisecond), sumDur.Round(time.Millisecond))
		return nil
	},
}

func shortenURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	path := u.Path
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	if len(path) > 70 {
		path = path[:67] + "..."
	}
	return u.Host + path
}

func init() {
	sourcetestCmd.Flags().StringVar(&sourcetestProxy, "proxy", "", "override proxy for this test (use 'none' to disable)")
	rootCmd.AddCommand(sourcetestCmd)
}
