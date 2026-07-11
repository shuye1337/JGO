package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"jgo/internal/env"
	"jgo/internal/wrapper"

	"github.com/spf13/cobra"
)

var mavenCmd = &cobra.Command{
	Use:   "maven",
	Short: "Configure Maven settings",
	Long:  "Configure Maven settings (proxy, MAVEN_HOME).",
}

var mavenProxyCmd = &cobra.Command{
	Use:   "proxy [url]",
	Short: "Set, show, or remove the Maven wrapper proxy",
	Long: "Set, show, or remove the proxy for Maven Wrapper (user-level).\n" +
		"Configures <proxies> in ~/.m2/settings.xml.\n\n" +
		"  jgo maven proxy                      # show current proxy\n" +
		"  jgo maven proxy host:port            # set proxy\n" +
		"  jgo maven proxy user:pass@host:port  # set proxy with auth\n" +
		"  jgo maven proxy none                 # remove proxy",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			result, err := wrapper.GetMavenProxy()
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil
		}
		return wrapper.SetMavenProxy(args[0])
	},
}

var mavenPathCmd = &cobra.Command{
	Use:   "path [path]",
	Short: "Set or show MAVEN_HOME",
	Long: "Set or show the MAVEN_HOME environment variable.\n" +
		"Also adds %MAVEN_HOME%\\bin (or $MAVEN_HOME/bin) to PATH.\n\n" +
		"  jgo maven path             # show current MAVEN_HOME\n" +
		"  jgo maven path /path/to/dir  # set MAVEN_HOME and update PATH",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			val, err := env.GetEnvVar("MAVEN_HOME")
			if err != nil {
				return err
			}
			if val == "" {
				fmt.Println("MAVEN_HOME is not set.")
			} else {
				fmt.Println("MAVEN_HOME=" + val)
			}
			return nil
		}

		path := args[0]
		if err := env.SetEnvVar("MAVEN_HOME", path); err != nil {
			return err
		}

		binPath := getMavenBinPath(path)
		if err := env.AddToPath(binPath); err != nil {
			return fmt.Errorf("failed to add to PATH: %w", err)
		}

		fmt.Printf("MAVEN_HOME set to: %s\n", path)
		fmt.Printf("Added to PATH: %s\n", binPath)
		return nil
	},
}

func getMavenBinPath(mavenHome string) string {
	return filepath.Join(mavenHome, "bin")
}

func userHomeDir() (string, error) {
	return os.UserHomeDir()
}

func init() {
	mavenCmd.AddCommand(mavenProxyCmd)
	mavenCmd.AddCommand(mavenPathCmd)
	rootCmd.AddCommand(mavenCmd)
}
