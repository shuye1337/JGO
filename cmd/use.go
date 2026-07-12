package cmd

import (
	"fmt"

	"jgo/internal/config"
	"jgo/internal/jdk"

	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Switch the active JDK (sets JAVA_HOME and PATH)",
	Long: "Switch the active JDK by name or version.\n" +
		"Sets JAVA_HOME and updates PATH system environment variables.\n\n" +
		"  jgo use temurin-21   # switch by exact name\n" +
		"  jgo use 21           # switch by version (if unique)\n" +
		"  jgo use              # interactive selection",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		mgr := jdk.NewManager(cfg)

		installed := mgr.ListInstalled()

		if len(args) == 0 {
			name, err := selectInstalledJDK("Select JDK to use: ", installed, cfg.Current)
			if err != nil {
				return err
			}
			return mgr.Use(name)
		}

		arg := args[0]

		if jdk, err := mgr.FindByName(arg); err == nil {
			return mgr.Use(jdk.Name)
		}

		matches := mgr.FindByVersion(arg)
		if len(matches) == 0 {
			return fmt.Errorf("no JDK found with name or version '%s'", arg)
		}
		if len(matches) == 1 {
			return mgr.Use(matches[0].Name)
		}

		var options []string
		for _, j := range matches {
			options = append(options, fmt.Sprintf("%s (%s, %s)", j.Name, j.Version, j.Source))
		}
		idx, err := selectFromMenu("Multiple JDKs match, select one: ", options)
		if err != nil {
			return err
		}
		return mgr.Use(matches[idx].Name)
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
