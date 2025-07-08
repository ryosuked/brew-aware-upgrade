package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	highestPriority = "highest_priority"
	priority        = "priority"
	largeSize       = "large_size"

	categoryNames = []string{highestPriority, priority, largeSize}
	dryRun        bool
	verbose       bool
)

// Categories represents the package categories in the YAML file
type Categories struct {
	HighestPriority []string `yaml:"highest_priority"`
	Priority        []string `yaml:"priority"`
	LargeSize       []string `yaml:"large_size"`
}

// CategoriesWrapper is the top-level structure of the YAML config
type CategoriesWrapper struct {
	Categories Categories `yaml:"categories"`
}

// logVerbose prints messages only if verbose mode is enabled
func logVerbose(format string, args ...interface{}) {
	if verbose {
		fmt.Printf("[VERBOSE] "+format+"\n", args...)
	}
}

// runCommand executes a system command and returns the output
func runCommand(cmd string, args ...string) string {
	logVerbose("Executing: %s %v", cmd, args)
	if dryRun {
		return ""
	}

	command := exec.Command(cmd, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing %s %v: %v\n", cmd, args, err)
	}
	return string(output)
}

// getOutdatedPackages retrieves the list of outdated Homebrew packages
func getOutdatedPackages() map[string]bool {
	command := exec.Command("brew", "outdated", "--quiet", "--greedy")
	output, err := command.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing brew outdated: %v\n", err)
		return map[string]bool{} // 空のマップを返す
	}

	outdated := strings.Split(strings.TrimSpace(string(output)), "\n")
	outdatedMap := make(map[string]bool)
	for _, pkg := range outdated {
		if pkg != "" {
			outdatedMap[pkg] = true
		}
	}

	logVerbose("Outdated packages: %v", outdatedMap)
	return outdatedMap
}

// findConfigFile searches for the configuration file in multiple locations
func findConfigFile(filename string) (string, error) {
	searchPaths := []string{
		filepath.Join(filepath.Dir(os.Args[0]), filename),                 // 実行ファイルと同じフォルダ
		filepath.Join(os.Getenv("HOME"), ".brew-aware-upgrade", filename), // ユーザーホームの .brew-aware-upgrade フォルダ
		filename, // 現在のフォルダ
	}

	if customPaths := os.Getenv("BREWUP_CONFIG_PATHS"); customPaths != "" {
		for _, path := range strings.Split(customPaths, ":") {
			searchPaths = append([]string{filepath.Join(path, filename)}, searchPaths...)
		}
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("config file not found: %s", filename)
}

// upgradeCategory upgrades packages in a specific category
func upgradeCategory(categoryName string, packages []string, outdatedPackages map[string]bool) {
	var upgradablePackages []string
	for _, pkg := range packages {
		if outdatedPackages[pkg] {
			upgradablePackages = append(upgradablePackages, pkg)
		}
	}
	if len(upgradablePackages) > 0 {
		fmt.Printf("Upgrading %s packages: %v\n", categoryName, upgradablePackages)
		runCommand("brew", append([]string{"upgrade"}, upgradablePackages...)...)
		runCommand("brew", "cleanup", "--prune=all")
	} else {
		logVerbose("No outdated packages found in category: %s", categoryName)
	}
}

// loadConfig loads the YAML configuration file
func loadConfig(filename string) (Categories, error) {
	configFile, err := findConfigFile(filename)
	if err != nil {
		return Categories{}, err
	}

	fmt.Printf("Using config file: %s\n", configFile)

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return Categories{}, fmt.Errorf("error reading YAML file: %v", err)
	}

	logVerbose("Config file content:\n%s", string(data))

	var wrapper CategoriesWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return Categories{}, fmt.Errorf("error parsing YAML: %v", err)
	}

	return wrapper.Categories, nil
}

// parseFlags parses command-line flags
func parseFlags() (map[string]bool, string) {
	categoryFlag := flag.String("c", "", "Specify categories to upgrade (comma-separated)")
	priorityFlag := flag.Bool("P", false, "Upgrade highest_priority and priority packages")
	dryRunFlag := flag.Bool("D", false, "Dry run mode: show commands without executing")
	verboseFlag := flag.Bool("v", false, "Enable verbose logging")
	helpFlag := flag.Bool("h", false, "Show help message")

	flag.Parse()
	dryRun = *dryRunFlag
	verbose = *verboseFlag

	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	selectedCategories := map[string]bool{}
	if *priorityFlag {
		selectedCategories[highestPriority] = true
		selectedCategories[priority] = true
	}
	if *categoryFlag != "" {
		for _, cat := range strings.Split(*categoryFlag, ",") {
			selectedCategories[strings.TrimSpace(cat)] = true
		}
	}

	// No specific categories selected -> Upgrade all
	if len(selectedCategories) == 0 {
		for _, cat := range categoryNames {
			selectedCategories[cat] = true
		}
	}

	configFilename := "packages.yaml"
	if len(flag.Args()) > 0 {
		configFilename = flag.Args()[0]
	}

	return selectedCategories, configFilename
}

// executeUpgrade handles the upgrade process based on parsed flags
func executeUpgrade(selectedCategories map[string]bool, categories Categories) {
	outdatedPackages := getOutdatedPackages()

	if selectedCategories[highestPriority] {
		upgradeCategory(highestPriority, categories.HighestPriority, outdatedPackages)
	}
	if selectedCategories[priority] {
		upgradeCategory(priority, categories.Priority, outdatedPackages)
	}
	if selectedCategories[largeSize] {
		for _, pkg := range categories.LargeSize {
			if outdatedPackages[pkg] {
				fmt.Printf("Upgrading large_size package: %s\n", pkg)
				runCommand("brew", "upgrade", pkg)
				runCommand("brew", "cleanup", "--prune=all")
			} else {
				logVerbose("Skipping package (not outdated): %s", pkg)
			}
		}
	}

	// If all categories are selected, perform a full upgrade
	if len(selectedCategories) == len(categoryNames) && len(outdatedPackages) > 0 {
		fmt.Println("All categories selected, running full upgrade.")
		runCommand("brew", "upgrade")
		runCommand("brew", "cleanup", "--prune=all")
	} else {
		logVerbose("No outdated packages found for full upgrade.")
	}
}

func main() {
	selectedCategories, configFilename := parseFlags()
	categories, err := loadConfig(configFilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	executeUpgrade(selectedCategories, categories)
}
