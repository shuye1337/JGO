package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"

	"jgo/internal/config"
	"jgo/internal/jdk"
	"jgo/internal/mirror"
	"jgo/internal/provider"

	"github.com/spf13/cobra"
)

var installProxy string

var installCmd = &cobra.Command{
	Use:   "install [version]",
	Short: "Download and install a JDK",
	Long: "Download and install a JDK from available sources.\n\n" +
		"If a version is specified (e.g., 'jgo install 21'), shows available sources for that version.\n" +
		"If no version is specified, lists all available JDKs for selection.\n\n" +
		"  jgo install 21                       # install JDK 21 using configured proxy\n" +
		"  jgo install 21 --proxy http://...    # install JDK 21 using a temporary proxy\n" +
		"  jgo install 21 --proxy none          # install JDK 21 without proxy",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		mgr := jdk.NewManager(cfg)

		proxy := cfg.Proxy
		if cmd.Flags().Changed("proxy") {
			if strings.ToLower(installProxy) == "none" {
				proxy = ""
			} else {
				proxy = installProxy
			}
		}
		mgr.Config.Proxy = proxy

		version := ""
		if len(args) == 1 {
			version = args[0]
		}

		fmt.Fprintf(os.Stderr, "Fetching available JDKs...\n")
		if proxy != "" {
			fmt.Fprintf(os.Stderr, "Using proxy: %s\n", proxy)
		}

		mirrors := mirror.Resolve(cfg.Mirrors)
		for source, mb := range mirrors.Backends {
			if m, ok := mb.(mirror.Mirror); ok {
				fmt.Fprintf(os.Stderr, "  %s source: %s (%s)\n", source, m.ID(), m.DisplayName())
			}
		}

		spinner := StartSpinner("Searching")
		assets := mgr.FindAssets(version)
		spinner.Stop()
		if len(assets) == 0 {
			if version != "" {
				return fmt.Errorf("no JDK version '%s' found for your platform (%s/%s)", version, provider.MapOS(), provider.MapArch())
			}
			return fmt.Errorf("no JDKs available for your platform (%s/%s)", provider.MapOS(), provider.MapArch())
		}

		sort.Slice(assets, func(i, j int) bool {
			if assets[i].Major != assets[j].Major {
				return assets[i].Major > assets[j].Major
			}
			return assets[i].Source < assets[j].Source
		})

		var options []string
		for _, a := range assets {
			options = append(options, fmt.Sprintf("%-12s %s  [%s]", a.Source, a.Version, a.FileType))
		}

		idx, err := selectFromMenu("Select JDK to install: ", options)
		if err != nil {
			return err
		}

		selected := assets[idx]

		name := fmt.Sprintf("%s-%d", selected.Source, selected.Major)
		if _, exists := cfg.JDKs[name]; exists {
			overwrite, err := promptString(fmt.Sprintf("JDK '%s' already exists. Overwrite? (y/n): ", name))
			if err != nil {
				return err
			}
			if !strings.EqualFold(overwrite, "y") {
				return fmt.Errorf("installation cancelled")
			}
			if err := mgr.Remove(name); err != nil {
				return err
			}
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		go func() {
			<-sigCh
			fmt.Fprintln(os.Stderr, "\nCancelling...")
			cancel()
		}()

		return mgr.Install(ctx, selected, name)
	},
}

func init() {
	installCmd.Flags().StringVar(&installProxy, "proxy", "", "override proxy for this request (use 'none' to disable)")
	rootCmd.AddCommand(installCmd)
}
