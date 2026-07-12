package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"jgo/internal/config"
	"jgo/internal/jdk"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Add a JDK from a local archive or an existing JDK directory",
	Long: "Add a JDK from a local .zip/.tar.gz archive, or manage an existing\n" +
		"JDK directory already on disk.\n\n" +
		"The path is inspected automatically:\n" +
		"  - If it is a .zip/.tar.gz/.tgz file, it is extracted and installed under jgo's root.\n" +
		"  - If it is a directory containing a valid JDK (bin/java and bin/javac), jgo\n" +
		"    registers it in place without copying or moving files.\n\n" +
		"You will be prompted to enter a custom name for this JDK.",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		mgr := jdk.NewManager(cfg)

		path := args[0]

		name, err := promptString("Enter a name for this JDK: ")
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("name cannot be empty")
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

		if jdk.IsExistingJDKDir(path) {
			return mgr.AddExisting(ctx, path, name)
		}
		return mgr.AddLocal(ctx, path, name)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}