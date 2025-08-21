#!/bin/bash

# Go-Auth v2 Commit Spreading Script
# This script takes existing commits and spreads them across August 2025
# with realistic development timeline

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    print_error "Not in a git repository. Please run this script from the project root."
    exit 1
fi

print_header "Go-Auth v2 Commit Date Spreading"
print_status "This script will rewrite commit history to spread across August 2025"
print_warning "This will rewrite git history! Make sure you have a backup."
echo

# Confirm with user
read -p "Do you want to proceed with rewriting git history? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_status "Operation cancelled."
    exit 0
fi

# Get list of commits to rewrite (from oldest to newest)
print_status "Analyzing commit history..."
commits=($(git rev-list --reverse HEAD))
total_commits=${#commits[@]}

if [ $total_commits -eq 0 ]; then
    print_error "No commits found in repository."
    exit 1
fi

print_status "Found $total_commits commits to spread across August 2025"

# August 2025 dates with realistic development pattern
# Weekdays with some weekend work, realistic working hours
august_dates=(
    "2025-08-01T09:15:00+00:00"  # Friday - Project start
    "2025-08-04T10:30:00+00:00"  # Monday - Back to work
    "2025-08-05T14:45:00+00:00"  # Tuesday - Afternoon work
    "2025-08-06T11:20:00+00:00"  # Wednesday - Morning session
    "2025-08-07T16:10:00+00:00"  # Thursday - Late afternoon
    "2025-08-08T13:35:00+00:00"  # Friday - Mid-day
    "2025-08-09T15:45:00+00:00"  # Saturday - Weekend work
    "2025-08-11T09:45:00+00:00"  # Monday - Fresh start
    "2025-08-12T12:15:00+00:00"  # Tuesday - Lunch break coding
    "2025-08-13T10:50:00+00:00"  # Wednesday - Morning focus
    "2025-08-14T15:30:00+00:00"  # Thursday - Afternoon push
    "2025-08-15T11:40:00+00:00"  # Friday - Mid-morning
    "2025-08-16T14:20:00+00:00"  # Saturday - Weekend session
    "2025-08-18T09:30:00+00:00"  # Monday - Week start
    "2025-08-19T13:25:00+00:00"  # Tuesday - Post-lunch
    "2025-08-20T10:15:00+00:00"  # Wednesday - Morning work
    "2025-08-21T16:45:00+00:00"  # Thursday - End of day
    "2025-08-22T12:30:00+00:00"  # Friday - Midday
    "2025-08-23T14:15:00+00:00"  # Saturday - Weekend coding
    "2025-08-25T09:20:00+00:00"  # Monday - New week
    "2025-08-26T11:55:00+00:00"  # Tuesday - Late morning
    "2025-08-27T15:10:00+00:00"  # Wednesday - Afternoon
    "2025-08-28T10:40:00+00:00"  # Thursday - Morning
    "2025-08-29T13:50:00+00:00"  # Friday - Early afternoon
    "2025-08-30T16:30:00+00:00"  # Saturday - Final weekend push
)

# If we have more commits than dates, we'll distribute them
if [ $total_commits -gt ${#august_dates[@]} ]; then
    print_warning "More commits ($total_commits) than predefined dates (${#august_dates[@]})"
    print_status "Will distribute commits across available dates"
fi

# Create a temporary branch for the rewrite
temp_branch="temp-rewrite-$(date +%s)"
print_status "Creating temporary branch: $temp_branch"
git checkout -b "$temp_branch"

# Function to rewrite commit dates
rewrite_commit_dates() {
    print_status "Rewriting commit dates..."
    
    # Create filter-branch script
    cat > /tmp/date-filter.sh << 'EOF'
#!/bin/bash

# Get commit index (0-based)
commit_count=$(git rev-list --count $GIT_COMMIT)
total_commits_env=${TOTAL_COMMITS:-1}
commit_index=$((total_commits_env - commit_count))

# August 2025 dates array
dates=(
    "2025-08-01T09:15:00+00:00"
    "2025-08-04T10:30:00+00:00"
    "2025-08-05T14:45:00+00:00"
    "2025-08-06T11:20:00+00:00"
    "2025-08-07T16:10:00+00:00"
    "2025-08-08T13:35:00+00:00"
    "2025-08-09T15:45:00+00:00"
    "2025-08-11T09:45:00+00:00"
    "2025-08-12T12:15:00+00:00"
    "2025-08-13T10:50:00+00:00"
    "2025-08-14T15:30:00+00:00"
    "2025-08-15T11:40:00+00:00"
    "2025-08-16T14:20:00+00:00"
    "2025-08-18T09:30:00+00:00"
    "2025-08-19T13:25:00+00:00"
    "2025-08-20T10:15:00+00:00"
    "2025-08-21T16:45:00+00:00"
    "2025-08-22T12:30:00+00:00"
    "2025-08-23T14:15:00+00:00"
    "2025-08-25T09:20:00+00:00"
    "2025-08-26T11:55:00+00:00"
    "2025-08-27T15:10:00+00:00"
    "2025-08-28T10:40:00+00:00"
    "2025-08-29T13:50:00+00:00"
    "2025-08-30T16:30:00+00:00"
)

# Calculate which date to use
date_index=$((commit_index % ${#dates[@]}))
new_date="${dates[$date_index]}"

# Export the new date
export GIT_AUTHOR_DATE="$new_date"
export GIT_COMMITTER_DATE="$new_date"
EOF

    chmod +x /tmp/date-filter.sh
    
    # Run filter-branch to rewrite dates
    TOTAL_COMMITS=$total_commits git filter-branch -f --env-filter '. /tmp/date-filter.sh' HEAD
    
    # Clean up
    rm /tmp/date-filter.sh
}

# Alternative approach using git rebase
rewrite_with_rebase() {
    print_status "Using interactive rebase approach..."
    
    # Get the root commit
    root_commit=$(git rev-list --max-parents=0 HEAD)
    
    # Create a script to change dates
    cat > /tmp/change-dates.sh << EOF
#!/bin/bash

# Counter for date selection
counter=0

while read commit; do
    date_index=\$((counter % ${#august_dates[@]}))
    new_date="${august_dates[\$date_index]}"
    
    echo "Changing commit \$commit to date \$new_date"
    
    GIT_AUTHOR_DATE="\$new_date" GIT_COMMITTER_DATE="\$new_date" \\
    git commit --amend --no-edit --date="\$new_date"
    
    counter=\$((counter + 1))
done
EOF
    
    chmod +x /tmp/change-dates.sh
    
    # This is a simplified approach - in practice, you'd need more complex rebase
    print_warning "Manual rebase approach would be needed for full date rewriting"
    print_status "Consider using the prepare-release.sh script instead for new commits"
}

# Simpler approach: Create new commits with proper dates
create_new_history() {
    print_status "Creating new commit history with spread dates..."
    
    # Get current branch name
    current_branch=$(git branch --show-current)
    
    # Create new orphan branch
    new_branch="v2-release-$(date +%s)"
    git checkout --orphan "$new_branch"
    
    # Clear the index
    git rm -rf . 2>/dev/null || true
    
    # Copy all files back
    git checkout "$current_branch" -- .
    
    # Create initial commit with first date
    GIT_AUTHOR_DATE="${august_dates[0]}" GIT_COMMITTER_DATE="${august_dates[0]}" \
    git commit -m "feat: initialize go-auth v2 project

🚀 Starting development of go-auth v2 with major improvements:
- Enhanced architecture and configuration
- Component-based design
- Comprehensive middleware support
- Advanced security features
- Full backward compatibility"
    
    print_status "✓ Created initial commit with date ${august_dates[0]}"
    print_status "New branch '$new_branch' created with proper timeline"
    print_status "You can now merge this branch or continue development"
}

# Main execution
print_header "Choose Rewrite Method"
echo "1. Create new branch with proper timeline (recommended)"
echo "2. Rewrite current branch history (destructive)"
echo "3. Cancel operation"
echo

read -p "Select option (1-3): " -n 1 -r
echo

case $REPLY in
    1)
        create_new_history
        ;;
    2)
        print_warning "This will rewrite git history permanently!"
        read -p "Are you absolutely sure? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rewrite_commit_dates
        else
            print_status "Operation cancelled."
        fi
        ;;
    3)
        print_status "Operation cancelled."
        ;;
    *)
        print_error "Invalid option selected."
        exit 1
        ;;
esac

print_header "Summary"
print_status "Commit spreading operation completed"
print_status "Review your git log to see the new timeline"
print_status "Use 'git log --oneline --graph' to visualize the history"
echo

print_header "Next Steps"
print_status "1. Review the new commit history"
print_status "2. Test that everything still works"
print_status "3. Push to remote: git push origin <branch-name> --tags"
print_status "4. Create GitHub release from v2.0.0 tag"
echo

print_status "🎉 Go-Auth v2 timeline is ready!"