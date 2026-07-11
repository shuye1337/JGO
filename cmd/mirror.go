package cmd

import (
	"fmt"
	"strings"

	"jgo/internal/config"
	"jgo/internal/mirror"
	"jgo/internal/provider"

	"github.com/spf13/cobra"
)

var mirrorCmd = &cobra.Command{
	Use:   "mirror [source] [mirror-id]",
	Short: "Set, show, or remove mirror for a JDK source",
	Long: "Set, show, or remove mirror for a JDK source.\n\n" +
		"  jgo mirror                          # show current mirrors and available plugins\n" +
		"  jgo mirror Temurin tsinghua         # use Tsinghua TUNA mirror for Temurin\n" +
		"  jgo mirror Temurin none             # remove mirror for Temurin",
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return showMirrors(cfg)
		}

		if len(args) == 1 {
			return fmt.Errorf("usage: jgo mirror <source> <mirror-id|none>")
		}

		sourceInput := args[0]
		mirrorInput := args[1]

		sourceName := normalizeSourceName(sourceInput)
		if sourceName == "" {
			return fmt.Errorf("unknown source: %s", sourceInput)
		}

		if strings.ToLower(mirrorInput) == "none" {
			delete(cfg.Mirrors, sourceName)
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf("Mirror removed for %s.\n", sourceName)
			return nil
		}

		m, ok := mirror.Get(mirrorInput)
		if !ok {
			return fmt.Errorf("unknown mirror: %s. Available: %s", mirrorInput, availableMirrorIDs())
		}
		if !mirrorSupportsSource(m, sourceName) {
			return fmt.Errorf("mirror %s does not support source %s", m.ID(), sourceName)
		}

		if cfg.Mirrors == nil {
			cfg.Mirrors = make(map[string]string)
		}
		cfg.Mirrors[sourceName] = m.ID()
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("%s mirror set to: %s (%s)\n", sourceName, m.ID(), m.DisplayName())
		return nil
	},
}

func showMirrors(cfg *config.Config) error {
	fmt.Println("Current mirrors:")
	if len(cfg.Mirrors) == 0 {
		fmt.Println("  (none)")
	} else {
		for source, mirrorID := range cfg.Mirrors {
			m, ok := mirror.Get(mirrorID)
			if ok {
				fmt.Printf("  %s -> %s (%s)\n", source, m.ID(), m.DisplayName())
			} else {
				fmt.Printf("  %s -> %s (not found)\n", source, mirrorID)
			}
		}
	}

	fmt.Println("\nAvailable mirror plugins:")
	plugins := mirror.All()
	if len(plugins) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, m := range plugins {
			fmt.Printf("  %s (%s) - supports: %s\n", m.ID(), m.DisplayName(), strings.Join(m.SupportedSources(), ", "))
		}
	}
	return nil
}

func normalizeSourceName(input string) string {
	for _, p := range provider.All(provider.MirrorSet{}) {
		if strings.EqualFold(p.Name(), input) {
			return p.Name()
		}
	}
	return ""
}

func availableMirrorIDs() string {
	plugins := mirror.All()
	ids := make([]string, 0, len(plugins))
	for _, m := range plugins {
		ids = append(ids, m.ID())
	}
	return strings.Join(ids, ", ")
}

func mirrorSupportsSource(m mirror.Mirror, source string) bool {
	for _, s := range m.SupportedSources() {
		if s == source {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(mirrorCmd)
}
