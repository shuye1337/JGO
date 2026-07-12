package cmd

import (
	"github.com/spf13/cobra"
)

const customUsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Use 25 }} {{.Short}}{{if .HasAvailableSubCommands}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
    {{rpad (printf "%s %s" .Parent.Name .Use) 23 }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

var rootCmd = &cobra.Command{
	Use:   "jgo",
	Short: "Java version manager - manage multiple JDK installations",
	Long: "jgo is a CLI tool for managing multiple JDK installations.\n\n" +
		"It supports downloading JDKs from multiple sources (Dragonwell, Corretto,\n" +
		"Azul Zulu, Temurin), installing from local archives, switching JAVA_HOME,\n" +
		"and configuring proxy settings for downloads and build tools.",
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SilenceErrors = true
}

func Execute() error {
	rootCmd.SetUsageTemplate(customUsageTemplate)
	return rootCmd.Execute()
}
