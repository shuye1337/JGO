package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"jgo/internal/config"
)

func selectFromMenu(prompt string, options []string) (int, error) {
	if len(options) == 0 {
		return -1, fmt.Errorf("no options available")
	}

	fmt.Println()
	for i, opt := range options {
		fmt.Printf("  [%d] %s\n", i+1, opt)
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			return -1, err
		}
		input = strings.TrimSpace(input)
		n, err := strconv.Atoi(input)
		if err != nil || n < 1 || n > len(options) {
			fmt.Fprintf(os.Stderr, "  Invalid choice, please enter 1-%d\n", len(options))
			continue
		}
		return n - 1, nil
	}
}

func selectInstalledJDK(prompt string, installed []config.JDKInfo, current string) (string, error) {
	if len(installed) == 0 {
		return "", fmt.Errorf("no JDKs installed. Use 'jgo install [version]' or 'jgo add <path>' first")
	}
	var options []string
	for _, j := range installed {
		marker := "  "
		if j.Name == current {
			marker = "* "
		}
		options = append(options, fmt.Sprintf("%s%s (%s, %s)", marker, j.Name, j.Version, j.Source))
	}
	idx, err := selectFromMenu(prompt, options)
	if err != nil {
		return "", err
	}
	return installed[idx].Name, nil
}

func promptYesNo(prompt string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" {
			return true, nil
		}
		if input == "n" || input == "no" {
			return false, nil
		}
		fmt.Fprintln(os.Stderr, "  Please enter 'y' or 'n'")
	}
}

func promptString(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}
