// Package main provides the unified Mole CLI binary.
// This replaces both `mole` (bash) and `mole.ps1` (PowerShell).
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tw93/mole/cmd/platform"
	"github.com/tw93/mole/internal/check"
	"github.com/tw93/mole/internal/clean"
	"github.com/tw93/mole/internal/installer"
	"github.com/tw93/mole/internal/manage"
	"github.com/tw93/mole/internal/optimize"
	"github.com/tw93/mole/internal/purge"
	"github.com/tw93/mole/internal/uninstall"
	"github.com/tw93/mole/pkg/config"
	"github.com/tw93/mole/pkg/logutil"
	"github.com/tw93/mole/pkg/ui"
)

const (
	version   = "1.0.0"
	buildTime = ""
)

func main() {
	if len(os.Args) < 2 {
		showInteractiveMenu()
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "clean":
		cmdClean(args)
	case "uninstall":
		cmdUninstall(args)
	case "optimize":
		cmdOptimize(args)
	case "analyze":
		cmdAnalyze(args)
	case "status":
		cmdStatus(args)
	case "purge":
		cmdPurge(args)
	case "installer":
		cmdInstaller(args)
	case "check":
		cmdCheck(args)
	case "touchid":
		cmdTouchID(args)
	case "completion":
		cmdCompletion(args)
	case "update":
		cmdUpdate(args)
	case "remove":
		cmdRemove(args)
	case "--version", "-v", "version":
		showVersion()
	case "--help", "-h", "help":
		showHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		showHelp()
		os.Exit(1)
	}
}

func showInteractiveMenu() {
	// For now, show text-based menu. Will be replaced with Bubble Tea TUI.
	fmt.Println(ui.Banner())
	fmt.Println()

	osVersion := platform.Current.OSVersion()
	fmt.Printf("  %s | %s\n\n",
		ui.FormatGray(osVersion.Name),
		ui.FormatGray(osVersion.Version+" ("+osVersion.Arch+")"),
	)

	options := []ui.MenuOption{
		{Label: "clean", Description: "Deep system cleanup", Icon: ui.IconClean},
		{Label: "uninstall", Description: "Smart app uninstaller", Icon: ui.IconTrash},
		{Label: "optimize", Description: "System optimization", Icon: "⚡"},
		{Label: "analyze", Description: "Disk space analyzer", Icon: ui.IconFolder},
		{Label: "status", Description: "System monitor", Icon: "◉"},
		{Label: "purge", Description: "Clean project artifacts", Icon: ui.IconList},
		{Label: "installer", Description: "Remove installer files", Icon: "⬇"},
		{Label: "check", Description: "System health checks", Icon: "✓"},
		{Label: "update", Description: "Update Mole", Icon: "↗"},
		{Label: "remove", Description: "Remove Mole from system", Icon: ui.IconTrash},
	}

	menu := ui.NewMenu("What would you like to do?", options)
	fmt.Println(menu.View())
	fmt.Println("\nRun: mo <command>  (e.g., mo clean)")
}

func showVersion() {
	fmt.Printf("Mole v%s", version)
	if buildTime != "" {
		fmt.Printf(" (built %s)", buildTime)
	}
	fmt.Println()

	osVersion := platform.Current.OSVersion()
	fmt.Printf("Platform: %s %s (%s)\n", osVersion.Name, osVersion.Version, osVersion.Arch)
}

func showHelp() {
	fmt.Println(ui.Banner())
	fmt.Println()
	fmt.Println("  Deep clean and optimize your system.")
	fmt.Println()
	fmt.Println("  " + ui.CyanStyle.Render("COMMANDS:"))
	fmt.Println()
	fmt.Println("    " + ui.CyanStyle.Render("clean") + "       Deep system cleanup")
	fmt.Println("    " + ui.CyanStyle.Render("uninstall") + "   Smart application uninstaller")
	fmt.Println("    " + ui.CyanStyle.Render("optimize") + "    System optimization and repairs")
	fmt.Println("    " + ui.CyanStyle.Render("analyze") + "     Disk space analyzer")
	fmt.Println("    " + ui.CyanStyle.Render("status") + "      System monitor")
	fmt.Println("    " + ui.CyanStyle.Render("purge") + "       Clean project artifacts")
	fmt.Println("    " + ui.CyanStyle.Render("installer") + "   Find and remove installer files")
	fmt.Println("    " + ui.CyanStyle.Render("check") + "       System health checks")
	fmt.Println("    " + ui.CyanStyle.Render("touchid") + "     Configure Touch ID / Touch ID for sudo")
	fmt.Println("    " + ui.CyanStyle.Render("completion") + "  Set up shell tab completion")
	fmt.Println("    " + ui.CyanStyle.Render("update") + "      Update Mole")
	fmt.Println("    " + ui.CyanStyle.Render("remove") + "      Remove Mole from system")
	fmt.Println()
	fmt.Println("  " + ui.CyanStyle.Render("OPTIONS:"))
	fmt.Println()
	fmt.Println("    " + ui.CyanStyle.Render("--version") + "   Show version information")
	fmt.Println("    " + ui.CyanStyle.Render("--help") + "      Show this help message")
	fmt.Println("    " + ui.CyanStyle.Render("--dry-run") + "   Preview without making changes")
	fmt.Println("    " + ui.CyanStyle.Render("--debug") + "     Enable debug logging")
	fmt.Println()
	fmt.Println("  " + ui.GrayStyle.Render("Run 'mo <command> --help' for command-specific help"))
}

// Command implementations.

func cmdClean(args []string) {
	logger := setupLogger(args)
	defer logger.Close()

	dryRun, _ := parseFlags(args)
	ctx := context.Background()
	cfg := config.New()

	err := clean.Run(ctx, cfg, logger, dryRun)
	if err != nil {
		logger.Error("Clean failed: %v", err)
		os.Exit(1)
	}
}

func cmdUninstall(args []string) {
	logger := setupLogger(args)
	defer logger.Close()

	dryRun, _ := parseFlags(args)
	ctx := context.Background()

	apps, err := platform.Current.ListApps(ctx)
	if err != nil {
		logger.Error("Error listing apps: %v", err)
		os.Exit(1)
	}

	if len(args) == 0 {
		fmt.Printf("Found %d installed applications\n\n", len(apps))
		for _, app := range apps {
			running := ""
			if app.IsRunning {
				running = " " + ui.FormatWarning("Running")
			}
			size := ""
			if app.Size > 0 {
				size = " " + ui.FormatGray(ui.FormatBytes(app.Size))
			}
			fmt.Printf("  %s %s%s%s\n", ui.IconArrow, app.Name, size, running)
		}
		fmt.Println("\nTo uninstall, pass the app name, e.g.: mo uninstall \"App Name\"")
		return
	}

	targetName := args[0]
	var targetApp *platform.AppInfo
	for _, app := range apps {
		if strings.EqualFold(app.Name, targetName) {
			targetApp = &app
			break
		}
	}

	if targetApp == nil {
		logger.Error("Application not found: %s", targetName)
		os.Exit(1)
	}

	uninstaller := uninstall.New(logger)
	err = uninstaller.Execute(ctx, *targetApp, dryRun)
	if err != nil {
		logger.Error("Uninstall failed: %v", err)
		os.Exit(1)
	}
}

func cmdOptimize(args []string) {
	logger := setupLogger(args)
	defer logger.Close()

	dryRun, _ := parseFlags(args)
	ctx := context.Background()

	optimizer := optimize.New(logger)
	err := optimizer.Run(ctx, dryRun)
	if err != nil {
		logger.Error("Optimize failed: %v", err)
		os.Exit(1)
	}
}

func cmdAnalyze(args []string) {
	// Delegate to cmd/analyze binary.
	fmt.Println(ui.FormatInfo("Analyze delegates to the Go TUI binary"))
	fmt.Println()
	fmt.Println("  Usage: mo analyze [path]")
	fmt.Println("  Default path: home directory")
	fmt.Println()
	fmt.Println(ui.FormatGray("  The analyze TUI binary must be built alongside the main binary."))
}

func cmdStatus(args []string) {
	// Delegate to cmd/status binary.
	fmt.Println(ui.FormatInfo("Status delegates to the Go TUI binary"))
	fmt.Println()
	fmt.Println("  Usage: mo status")
	fmt.Println()
	fmt.Println(ui.FormatGray("  The status TUI binary must be built alongside the main binary."))
}

func cmdPurge(args []string) {
	logger := setupLogger(args)
	defer logger.Close()

	dryRun, _ := parseFlags(args)
	ctx := context.Background()

	purger := purge.New(logger)
	err := purger.Run(ctx, dryRun)
	if err != nil {
		logger.Error("Purge failed: %v", err)
		os.Exit(1)
	}
}

func cmdInstaller(args []string) {
	logger := setupLogger(args)
	defer logger.Close()

	dryRun, _ := parseFlags(args)
	ctx := context.Background()

	finder := installer.New(logger)
	err := finder.Run(ctx, dryRun)
	if err != nil {
		logger.Error("Installer cleanup failed: %v", err)
		os.Exit(1)
	}
}

func cmdCheck(args []string) {
	logger := setupLogger(args)
	defer logger.Close()

	ctx := context.Background()
	checker := check.New(logger)
	err := checker.Run(ctx)
	if err != nil {
		logger.Error("Check failed: %v", err)
		os.Exit(1)
	}
}

func cmdTouchID(args []string) {
	// Touch ID is macOS-only. On Windows, show Windows Hello info.
	osVersion := platform.Current.OSVersion()
	if osVersion.Name == "Windows" {
		fmt.Println(ui.FormatInfo("Windows Hello is the Windows equivalent of Touch ID"))
		fmt.Println()
		fmt.Println("  Windows Hello supports:")
		fmt.Println("  " + ui.FormatGray("• Face recognition"))
		fmt.Println("  " + ui.FormatGray("• Fingerprint authentication"))
		fmt.Println("  " + ui.FormatGray("• PIN authentication"))
		fmt.Println()
		fmt.Println("  Mole uses UAC elevation for admin operations on Windows.")
		return
	}

	fmt.Println(ui.FormatInfo("Touch ID configuration is macOS-only."))
}

func cmdCompletion(args []string) {
	logger := setupLogger(args)
	defer logger.Close()

	shell := "bash"
	if len(args) > 0 {
		shell = args[0]
	}
	dryRun, _ := parseFlags(args)

	err := manage.New(logger).Completion(shell, dryRun)
	if err != nil {
		logger.Error("Completion setup failed: %v", err)
		os.Exit(1)
	}
}

func cmdUpdate(args []string) {
	logger := setupLogger(args)
	defer logger.Close()

	err := manage.New(logger).Update()
	if err != nil {
		logger.Error("Update failed: %v", err)
		os.Exit(1)
	}
}

func cmdRemove(args []string) {
	logger := setupLogger(args)
	defer logger.Close()

	err := manage.New(logger).Remove()
	if err != nil {
		logger.Error("Remove failed: %v", err)
		os.Exit(1)
	}
}

// Flag parsing helpers.

func parseFlags(args []string) (dryRun, debug bool) {
	for _, arg := range args {
		switch arg {
		case "--dry-run", "-n":
			dryRun = true
		case "--debug", "-d":
			debug = true
		}
	}
	return
}

func setupLogger(args []string) *logutil.Logger {
	_, debug := parseFlags(args)
	logDir := platform.Current.MoleLogDir()

	var cfg logutil.Config
	cfg.LogDir = logDir
	cfg.DebugMode = debug
	cfg.OpLog = true

	logger, err := logutil.New(cfg)
	if err != nil {
		// Fall back to console logger
		return logutil.NewConsoleLogger()
	}
	return logger
}
