package cmd

import (
	"fmt"

	"jgo/internal/config"

	"github.com/spf13/cobra"
)

var rootCmd2 = &cobra.Command{
	Use:   "root [path]",
	Short: "Set or show the JDK installation root path",
	Long: "Set or show the JDK installation root path.\n\n" +
		"  jgo root              # show current root path\n" +
		"  jgo root /path/to/dir # set root path",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			if cfg.RootPath == "" {
				fmt.Println("Root path is not set.")
			} else {
				fmt.Println("Root path:", cfg.RootPath)
			}
			return nil
		}

		cfg.RootPath = args[0]
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Root path set to: %s\n", cfg.RootPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rootCmd2)
}
