package auth

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CodeMigrationTool provides automated code migration assistance.
type CodeMigrationTool struct {
	patterns map[string]string
}

// NewCodeMigrationTool creates a new code migration tool.
func NewCodeMigrationTool() *CodeMigrationTool {
	return &CodeMigrationTool{
		patterns: map[string]string{
			// Service creation patterns
			`auth\.NewAuthService\(([^)]+)\)`: `auth.New($1) // MIGRATION: Consider using auth.NewSQLite() or auth.NewPostgres() for clearer intent`,
			
			// Registration patterns
			`\.Register\(auth\.RegisterPayload\{`: `.Register(auth.RegisterRequest{`,
			`RegisterPayload\{`: `RegisterRequest{`,
			
			// Login patterns
			`\*auth\.LoginResponse`: `*auth.LoginResult`,
			`LoginResponse\{`: `LoginResult{`,
			
			// Storage access patterns (suggest using new components)
			`storage\.GetUserByID\(([^)]+)\)`: `auth.Users().Get($1) // MIGRATION: Use Users component for safer access`,
			`storage\.GetUserByUsername\(([^)]+)\)`: `auth.Users().GetByUsername($1) // MIGRATION: Use Users component`,
			`storage\.GetUserByEmail\(([^)]+)\)`: `auth.Users().GetByEmail($1) // MIGRATION: Use Users component`,
			
			// Token validation patterns
			`\.ValidateAccessToken\(([^)]+)\)`: `.ValidateAccessToken($1) // MIGRATION: Consider using auth.Tokens().Validate() for enhanced features`,
			`\.ValidateRefreshToken\(([^)]+)\)`: `.ValidateRefreshToken($1) // MIGRATION: Consider using auth.Tokens().Refresh() for token refresh`,
		},
	}
}

// MigrateFile analyzes a Go file and suggests migrations.
func (t *CodeMigrationTool) MigrateFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	var lines []string
	var suggestions []string
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		lines = append(lines, line)
		
		// Check each pattern
		for pattern, replacement := range t.patterns {
			re := regexp.MustCompile(pattern)
			if re.MatchString(line) {
				suggestion := fmt.Sprintf("Line %d: %s", lineNum, line)
				suggestion += fmt.Sprintf("\n  Suggested: %s", re.ReplaceAllString(line, replacement))
				suggestions = append(suggestions, suggestion)
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	
	if len(suggestions) > 0 {
		fmt.Printf("Migration suggestions for %s:\n", filePath)
		fmt.Println(strings.Repeat("=", 50))
		for _, suggestion := range suggestions {
			fmt.Println(suggestion)
			fmt.Println()
		}
	} else {
		fmt.Printf("No migration suggestions needed for %s\n", filePath)
	}
	
	return nil
}

// GenerateMigrationScript creates a shell script to help with migration.
func (t *CodeMigrationTool) GenerateMigrationScript(outputPath string) error {
	script := `#!/bin/bash
# Go-Auth v2 Migration Script
# This script helps migrate from go-auth v1 to v2

echo "Go-Auth v2 Migration Assistant"
echo "=============================="
echo ""

# Check if go.mod exists
if [ ! -f "go.mod" ]; then
    echo "Error: go.mod not found. Please run this script from your Go project root."
    exit 1
fi

echo "1. Updating go.mod to use go-auth v2..."
go get github.com/pragneshbagary/go-auth@latest

echo ""
echo "2. Scanning for migration opportunities..."
echo ""

# Find all Go files and check for old patterns
find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | while read -r file; do
    echo "Checking $file..."
    
    # Check for old AuthService usage
    if grep -q "NewAuthService\|RegisterPayload\|LoginResponse" "$file"; then
        echo "  ⚠️  Found v1 API usage in $file"
        echo "     Consider updating to v2 API. See migration guide for details."
    fi
    
    # Check for direct storage access
    if grep -q "storage\.GetUser\|storage\.CreateUser" "$file"; then
        echo "  ℹ️  Found direct storage access in $file"
        echo "     Consider using auth.Users() component for safer access."
    fi
done

echo ""
echo "3. Migration complete! Next steps:"
echo "   - Review the migration guide: https://github.com/pragneshbagary/go-auth/blob/main/MIGRATION.md"
echo "   - Update your code to use the new v2 API"
echo "   - Test your application thoroughly"
echo "   - Consider using the new component-based API (Users, Tokens, Middleware)"
echo ""
echo "For detailed migration assistance, use the Go migration tool:"
echo "   go run -c 'import \"github.com/pragneshbagary/go-auth/pkg/auth\"; auth.NewCodeMigrationTool().MigrateFile(\"your-file.go\")'"
`

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create migration script: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(script)
	if err != nil {
		return fmt.Errorf("failed to write migration script: %w", err)
	}

	// Make the script executable
	err = os.Chmod(outputPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	return nil
}

// AnalyzeProject scans an entire project for migration opportunities.
func (t *CodeMigrationTool) AnalyzeProject(projectPath string) (*ProjectMigrationReport, error) {
	report := &ProjectMigrationReport{
		ProjectPath: projectPath,
		Files:       make(map[string][]string),
	}

	// Walk through all Go files in the project
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and .git directories
		if strings.Contains(path, "vendor/") || strings.Contains(path, ".git/") {
			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Analyze the file
		suggestions, err := t.analyzeFile(path)
		if err != nil {
			return err
		}

		if len(suggestions) > 0 {
			report.Files[path] = suggestions
			report.TotalFiles++
			report.TotalSuggestions += len(suggestions)
		}

		return nil
	})

	return report, err
}

// analyzeFile analyzes a single file and returns suggestions.
func (t *CodeMigrationTool) analyzeFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var suggestions []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check each pattern
		for pattern, replacement := range t.patterns {
			re := regexp.MustCompile(pattern)
			if re.MatchString(line) {
				suggestion := fmt.Sprintf("Line %d: %s -> %s", lineNum, strings.TrimSpace(line), re.ReplaceAllString(line, replacement))
				suggestions = append(suggestions, suggestion)
			}
		}
	}

	return suggestions, scanner.Err()
}

// ProjectMigrationReport contains the results of a project-wide migration analysis.
type ProjectMigrationReport struct {
	ProjectPath      string
	Files            map[string][]string
	TotalFiles       int
	TotalSuggestions int
}

// Print outputs the migration report to stdout.
func (r *ProjectMigrationReport) Print() {
	fmt.Printf("Go-Auth v2 Migration Report for %s\n", r.ProjectPath)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Files with migration opportunities: %d\n", r.TotalFiles)
	fmt.Printf("Total suggestions: %d\n\n", r.TotalSuggestions)

	if r.TotalFiles == 0 {
		fmt.Println("✅ No migration needed! Your code appears to be compatible with v2.")
		return
	}

	for filePath, suggestions := range r.Files {
		fmt.Printf("📁 %s\n", filePath)
		for _, suggestion := range suggestions {
			fmt.Printf("   %s\n", suggestion)
		}
		fmt.Println()
	}

	fmt.Println("Migration Guide: https://github.com/pragneshbagary/go-auth/blob/main/MIGRATION.md")
}

// Save saves the migration report to a file.
func (r *ProjectMigrationReport) Save(outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	fmt.Fprintf(writer, "Go-Auth v2 Migration Report for %s\n", r.ProjectPath)
	fmt.Fprintf(writer, "%s\n", strings.Repeat("=", 60))
	fmt.Fprintf(writer, "Files with migration opportunities: %d\n", r.TotalFiles)
	fmt.Fprintf(writer, "Total suggestions: %d\n\n", r.TotalSuggestions)

	if r.TotalFiles == 0 {
		fmt.Fprintf(writer, "✅ No migration needed! Your code appears to be compatible with v2.\n")
		return nil
	}

	for filePath, suggestions := range r.Files {
		fmt.Fprintf(writer, "📁 %s\n", filePath)
		for _, suggestion := range suggestions {
			fmt.Fprintf(writer, "   %s\n", suggestion)
		}
		fmt.Fprintf(writer, "\n")
	}

	fmt.Fprintf(writer, "Migration Guide: https://github.com/pragneshbagary/go-auth/blob/main/MIGRATION.md\n")
	return nil
}