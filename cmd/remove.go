package cmd

import (
	"fmt"
	"os"

	"jgo/internal/config"
	"jgo/internal/jdk"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a JDK from managed JDKs",
	Long: "Remove a JDK from the managed JDKs list.\n" +
		"You will be asked whether to also delete the JDK folder from disk.\n\n" +
		"  jgo remove temurin-21   # remove by exact name\n" +
		"  jgo remove              # interactive selection",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		mgr := jdk.NewManager(cfg)

		name := ""
		if len(args) == 0 {
			installed := mgr.ListInstalled()
			n, err := selectInstalledJDK("Select JDK to remove: ", installed, cfg.Current)
			if err != nil {
				return err
			}
			name = n
		} else {
			name = args[0]
		}

		jdkInfo, err := mgr.FindByName(name)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "JDK: %s (%s, %s)\n", jdkInfo.Name, jdkInfo.Version, jdkInfo.Source)
		fmt.Fprintf(os.Stderr, "Path: %s\n", jdkInfo.Path)
		if cfg.Current == name {
			fmt.Fprintln(os.Stderr, "This is currently the active JDK.")
		}

		name = jdkInfo.Name

		fmt.Fprintln(os.Stderr)
		confirm, err := promptYesNo(fmt.Sprintf("Remove '%s' from managed JDKs? (y/n): ", name))
		if err != nil {
			return err
		}
		if !confirm {
			return fmt.Errorf("removal cancelled")
		}

		deleteDir := false
		if jdkInfo.Source == "external" {
			fmt.Fprintf(os.Stderr, "JDK is managed in place at %s; only the jgo registry entry will be removed.\n", jdkInfo.Path)
		} else if _, err := os.Stat(jdkInfo.Path); err == nil {
			deleteDir, err = promptYesNo(fmt.Sprintf("Also delete the JDK folder at %s? This is IRREVERSIBLE. (y/n): ", jdkInfo.Path))
			if err != nil {
				return err
			}
		} else {
			fmt.Fprintf(os.Stderr, "Directory does not exist, skipped.\n")
		}

		return mgr.Remove(name, deleteDir)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
