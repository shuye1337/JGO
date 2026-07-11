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
	Short: "Add a local JDK from a .zip or .tar.gz archive",
	Long: "Add a local JDK from a .zip or .tar.gz archive.\n" +
		"The archive must contain a valid JDK (with bin/java and bin/javac).\n" +
		"You will be prompted to enter a custom name for this JDK.",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		mgr := jdk.NewManager(cfg)

		archivePath := args[0]

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

		return mgr.AddLocal(ctx, archivePath, name)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
