package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"jgo/internal/config"
	"jgo/internal/jdk"
	"jgo/internal/mirror"
	"jgo/internal/provider"

	"github.com/spf13/cobra"
)

var listAvailable bool
var listProxy string

var listCmd = &cobra.Command{
	Use:   "list [available]",
	Short: "List installed JDKs or available JDKs",
	Long: "List installed JDKs.\n" +
		"Use 'jgo list available' to list JDKs available for download from all sources.\n\n" +
		"  jgo list available                    # list available JDKs using configured proxy\n" +
		"  jgo list available --proxy http://... # list available JDKs using a temporary proxy\n" +
		"  jgo list available --proxy none       # list available JDKs without proxy",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		mgr := jdk.NewManager(cfg)

		if len(args) == 1 && args[0] == "available" || listAvailable {
			proxy := cfg.Proxy
			if cmd.Flags().Changed("proxy") {
				if strings.ToLower(listProxy) == "none" {
					proxy = ""
				} else {
					proxy = listProxy
				}
			}
			return listAvailableJDKs(mgr, proxy)
		}
		return listInstalledJDKs(mgr)
	},
}

func init() {
	listCmd.Flags().BoolVar(&listAvailable, "available", false, "list available JDKs from all sources")
	listCmd.Flags().StringVar(&listProxy, "proxy", "", "override proxy for this request (use 'none' to disable)")
	rootCmd.AddCommand(listCmd)
}

func listInstalledJDKs(mgr *jdk.Manager) error {
	installed := mgr.ListInstalled()
	if len(installed) == 0 {
		fmt.Println("No JDKs installed. Use 'jgo install [version]' or 'jgo add <path>' to add one.")
		return nil
	}

	current := mgr.Config.Current
	fmt.Printf("%-20s %-20s %-12s %s\n", "Name", "Version", "Source", "Path")
	fmt.Println(strings.Repeat("-", 80))
	for _, j := range installed {
		marker := "  "
		if j.Name == current {
			marker = "* "
		}
		fmt.Printf("%s%-19s %-20s %-12s %s\n", marker, j.Name, j.Version, j.Source, j.Path)
	}
	return nil
}

func listAvailableJDKs(mgr *jdk.Manager, proxy string) error {
	osName := provider.MapOS()
	arch := provider.MapArch()
	if proxy != "" {
		fmt.Fprintf(os.Stderr, "Using proxy: %s\n", proxy)
	}

	mirrors := mirror.Resolve(mgr.Config.Mirrors)
	for source, mb := range mirrors.Backends {
		if m, ok := mb.(mirror.Mirror); ok {
			fmt.Fprintf(os.Stderr, "  %s source: %s (%s)\n", source, m.ID(), m.DisplayName())
		}
	}

	spinner := StartSpinner("Searching available JDKs")
	assets, errs := provider.ListAllAvailable(osName, arch, proxy, mirrors)
	spinner.Stop()

	fmt.Printf("Available JDKs for %s/%s:\n", osName, arch)
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "  Warning: %v\n", e)
	}

	if len(assets) == 0 {
		fmt.Println("  No JDKs found for your platform.")
		return nil
	}

	sort.Slice(assets, func(i, j int) bool {
		if assets[i].Source != assets[j].Source {
			return assets[i].Source < assets[j].Source
		}
		return assets[i].Major > assets[j].Major
	})

	groups := make(map[string][]provider.JDKAsset)
	for _, a := range assets {
		groups[a.Source] = append(groups[a.Source], a)
	}

	sources := make([]string, 0, len(groups))
	for s := range groups {
		sources = append(sources, s)
	}
	sort.Strings(sources)

	for _, src := range sources {
		fmt.Printf("\n%s:\n", titleCase(src))
		fmt.Printf("  %-6s %-20s %-8s\n", "Major", "Version", "Type")
		fmt.Printf("  %s\n", strings.Repeat("-", 40))
		for _, a := range groups[src] {
			fmt.Printf("  %-6d %-20s %s\n", a.Major, a.Version, a.FileType)
		}
	}

	return nil
}

func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
