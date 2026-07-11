package cmd

import (
	"fmt"
	"path/filepath"

	"jgo/internal/env"
	"jgo/internal/wrapper"

	"github.com/spf13/cobra"
)

var gradleCmd = &cobra.Command{
	Use:   "gradle",
	Short: "Configure Gradle settings",
	Long:  "Configure Gradle settings (proxy, GRADLE_USER_HOME).",
}

var gradleProxyCmd = &cobra.Command{
	Use:   "proxy [url]",
	Short: "Set, show, or remove the Gradle wrapper proxy",
	Long: "Set, show, or remove the proxy for Gradle Wrapper (user-level).\n" +
		"Configures systemProp.* in ~/.gradle/gradle.properties.\n\n" +
		"  jgo gradle proxy                      # show current proxy\n" +
		"  jgo gradle proxy host:port            # set proxy\n" +
		"  jgo gradle proxy user:pass@host:port  # set proxy with auth\n" +
		"  jgo gradle proxy none                 # remove proxy",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			result, err := wrapper.GetGradleProxy()
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil
		}
		return wrapper.SetGradleProxy(args[0])
	},
}

var gradlePathCmd = &cobra.Command{
	Use:   "path [path]",
	Short: "Set or show GRADLE_USER_HOME",
	Long: "Set or show the GRADLE_USER_HOME environment variable.\n" +
		"This is the Gradle cache/config directory (not the executable directory).\n\n" +
		"  jgo gradle path             # show current GRADLE_USER_HOME\n" +
		"  jgo gradle path /path/to/dir  # set GRADLE_USER_HOME",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			val, err := env.GetEnvVar("GRADLE_USER_HOME")
			if err != nil {
				return err
			}
			if val == "" {
				fmt.Println("GRADLE_USER_HOME is not set.")
				defaultPath, _ := defaultGradleUserHome()
				if defaultPath != "" {
					fmt.Printf("Default: %s\n", defaultPath)
				}
			} else {
				fmt.Println("GRADLE_USER_HOME=" + val)
			}
			return nil
		}

		path := args[0]
		if err := env.SetEnvVar("GRADLE_USER_HOME", path); err != nil {
			return err
		}
		fmt.Printf("GRADLE_USER_HOME set to: %s\n", path)
		return nil
	},
}

func defaultGradleUserHome() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gradle"), nil
}

func init() {
	gradleCmd.AddCommand(gradleProxyCmd)
	gradleCmd.AddCommand(gradlePathCmd)
	rootCmd.AddCommand(gradleCmd)
}
