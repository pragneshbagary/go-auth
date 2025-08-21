package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	var (
		projectPath = flag.String("path", ".", "Path to the Go project to analyze")
		outputFile  = flag.String("output", "", "Output file for the migration report (optional)")
		scriptPath  = flag.String("script", "", "Generate migration script at the specified path")
		fileToCheck = flag.String("file", "", "Analyze a specific Go file")
		showHelp    = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	if *showHelp {
		showUsage()
		return
	}

	tool := auth.NewCodeMigrationTool()

	// Generate migration script if requested
	if *scriptPath != "" {
		fmt.Printf("Generating migration script at %s...\n", *scriptPath)
		if err := tool.GenerateMigrationScript(*scriptPath); err != nil {
			log.Fatalf("Failed to generate migration script: %v", err)
		}
		fmt.Printf("✅ Migration script generated successfully!\n")
		fmt.Printf("Run: chmod +x %s && ./%s\n", *scriptPath, *scriptPath)
		return
	}

	// Analyze a specific file if requested
	if *fileToCheck != "" {
		fmt.Printf("Analyzing file: %s\n\n", *fileToCheck)
		if err := tool.MigrateFile(*fileToCheck); err != nil {
			log.Fatalf("Failed to analyze file: %v", err)
		}
		return
	}

	// Analyze entire project
	fmt.Printf("Analyzing project at: %s\n", *projectPath)
	fmt.Println("This may take a moment for large projects...")
	fmt.Println()

	report, err := tool.AnalyzeProject(*projectPath)
	if err != nil {
		log.Fatalf("Failed to analyze project: %v", err)
	}

	// Print report to stdout
	report.Print()

	// Save report to file if requested
	if *outputFile != "" {
		fmt.Printf("\nSaving report to %s...\n", *outputFile)
		if err := report.Save(*outputFile); err != nil {
			log.Fatalf("Failed to save report: %v", err)
		}
		fmt.Printf("✅ Report saved successfully!\n")
	}

	// Provide next steps
	if report.TotalFiles > 0 {
		fmt.Println("\n📋 Next Steps:")
		fmt.Println("1. Review the migration suggestions above")
		fmt.Println("2. Update your code to use the new v2 API")
		fmt.Println("3. Test your application thoroughly")
		fmt.Println("4. Consider using the new component-based API (Users, Tokens, Middleware)")
		fmt.Println("\n📖 Resources:")
		fmt.Println("- Migration Guide: https://github.com/pragneshbagary/go-auth/blob/main/MIGRATION.md")
		fmt.Println("- v2 Documentation: https://github.com/pragneshbagary/go-auth#readme")
	}
}

func showUsage() {
	fmt.Println("Go-Auth v2 Migration Tool")
	fmt.Println("========================")
	fmt.Println()
	fmt.Println("This tool helps you migrate from go-auth v1 to v2 by analyzing your code")
	fmt.Println("and providing specific migration suggestions.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  migrate [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -path string")
	fmt.Println("        Path to the Go project to analyze (default \".\")")
	fmt.Println("  -output string")
	fmt.Println("        Output file for the migration report (optional)")
	fmt.Println("  -script string")
	fmt.Println("        Generate migration script at the specified path")
	fmt.Println("  -file string")
	fmt.Println("        Analyze a specific Go file")
	fmt.Println("  -help")
	fmt.Println("        Show this help information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Analyze current directory")
	fmt.Println("  migrate")
	fmt.Println()
	fmt.Println("  # Analyze specific project and save report")
	fmt.Println("  migrate -path /path/to/project -output migration-report.txt")
	fmt.Println()
	fmt.Println("  # Analyze a specific file")
	fmt.Println("  migrate -file main.go")
	fmt.Println()
	fmt.Println("  # Generate migration script")
	fmt.Println("  migrate -script migrate.sh")
	fmt.Println()
	fmt.Println("For more information, visit:")
	fmt.Println("https://github.com/pragneshbagary/go-auth/blob/main/MIGRATION.md")
}