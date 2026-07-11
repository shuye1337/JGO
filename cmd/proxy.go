package cmd

import (
	"fmt"
	"strings"

	"jgo/internal/config"

	"github.com/spf13/cobra"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy [url]",
	Short: "Set, show, or remove the download proxy",
	Long: "Set, show, or remove the HTTP/HTTPS proxy used for JDK downloads.\n\n" +
		"  jgo proxy                # show current proxy\n" +
		"  jgo proxy http://host:port  # set proxy\n" +
		"  jgo proxy none           # remove proxy",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			if cfg.Proxy == "" {
				fmt.Println("No proxy configured.")
			} else {
				fmt.Println("Proxy:", cfg.Proxy)
			}
			return nil
		}

		proxyStr := args[0]
		if strings.ToLower(proxyStr) == "none" {
			proxyStr = ""
		}

		cfg.Proxy = proxyStr
		if err := cfg.Save(); err != nil {
			return err
		}

		if proxyStr == "" {
			fmt.Println("Proxy removed.")
		} else {
			fmt.Printf("Proxy set to: %s\n", proxyStr)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)
}
