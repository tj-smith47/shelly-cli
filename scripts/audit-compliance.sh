#!/usr/bin/env bash
#
# audit-compliance.sh - Enforce shelly-cli project rules from RULES.md
#
# Usage: ./scripts/audit-compliance.sh [--fix-summary]
#
# Exit codes:
#   0 - All checks pass
#   1 - Violations found
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Counters
ERRORS=0
WARNINGS=0

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Helper functions
error() {
    echo -e "${RED}ERROR${NC}: $1"
    ((ERRORS++)) || true
}

warning() {
    echo -e "${YELLOW}WARN${NC}: $1"
    ((WARNINGS++)) || true
}

info() {
    echo -e "${CYAN}INFO${NC}: $1"
}

section() {
    echo ""
    echo -e "${BOLD}=== $1 ===${NC}"
}

# Check for pattern in files and report violations
check_pattern() {
    local pattern="$1"
    local message="$2"
    local scope="$3"  # "cmd" or "all"
    local severity="$4"  # "error" or "warning"
    local exclude_pattern="${5:-}"

    local search_path
    if [[ "$scope" == "cmd" ]]; then
        search_path="internal/cmd"
    else
        search_path="internal"
    fi

    # Skip if path doesn't exist
    [[ ! -d "$search_path" ]] && return

    local results
    if [[ -n "$exclude_pattern" ]]; then
        results=$(grep -rn --include='*.go' -e "$pattern" "$search_path" 2>/dev/null | grep -v "$exclude_pattern" || true)
    else
        results=$(grep -rn --include='*.go' -e "$pattern" "$search_path" 2>/dev/null || true)
    fi

    if [[ -n "$results" ]]; then
        while IFS= read -r line; do
            if [[ "$severity" == "error" ]]; then
                error "$line - $message"
            else
                warning "$line - $message"
            fi
        done <<< "$results"
    fi
}

# Check for pattern using case-insensitive grep
check_pattern_nocase() {
    local pattern="$1"
    local message="$2"
    local scope="$3"
    local severity="$4"

    local search_path
    if [[ "$scope" == "cmd" ]]; then
        search_path="internal/cmd"
    else
        search_path="internal"
    fi

    [[ ! -d "$search_path" ]] && return

    local results
    results=$(grep -rin --include='*.go' -e "$pattern" "$search_path" 2>/dev/null || true)

    if [[ -n "$results" ]]; then
        while IFS= read -r line; do
            if [[ "$severity" == "error" ]]; then
                error "$line - $message"
            else
                warning "$line - $message"
            fi
        done <<< "$results"
    fi
}

# ============================================================================
# COMMAND-SPECIFIC RULES (internal/cmd/ only)
# ============================================================================

section "Command-Specific Rules (internal/cmd/)"

# Rule 1: Factory pattern required
# Find .go files in internal/cmd/ that define NewCommand but don't use factory pattern
info "Checking factory pattern..."
if [[ -d "internal/cmd" ]]; then
    # Find files with NewCommand that don't have the factory signature
    while IFS= read -r file; do
        # Skip test files and root.go
        [[ "$file" == *"_test.go" ]] && continue
        [[ "$file" == *"root.go" ]] && continue

        # Check if file has NewCommand function
        if grep -q "func NewCommand\|func NewCmd" "$file" 2>/dev/null; then
            # Check if it has factory parameter
            if ! grep -q "func NewCommand(f \*cmdutil\.Factory)" "$file" && \
               ! grep -q "func NewCmd.*\*cmdutil\.Factory" "$file"; then
                error "$file: Command missing factory pattern - func NewCommand(f *cmdutil.Factory)"
            fi
        fi
    done < <(find internal/cmd -name '*.go' -type f 2>/dev/null)
fi

# Rule 2: Commands must have Aliases
info "Checking for Aliases field..."
if [[ -d "internal/cmd" ]]; then
    while IFS= read -r file; do
        [[ "$file" == *"_test.go" ]] && continue
        [[ "$file" == *"root.go" ]] && continue

        # Check if file has cobra.Command definition
        if grep -q '&cobra\.Command{' "$file" 2>/dev/null; then
            # Check if it has Aliases field (allowing for multi-line definitions)
            if ! grep -q 'Aliases:' "$file"; then
                error "$file: Command missing Aliases field"
            fi
        fi
    done < <(find internal/cmd -name '*.go' -type f 2>/dev/null)
fi

# Rule 3: Commands must have Examples
info "Checking for Example field..."
if [[ -d "internal/cmd" ]]; then
    while IFS= read -r file; do
        [[ "$file" == *"_test.go" ]] && continue
        [[ "$file" == *"root.go" ]] && continue

        # Check if file has cobra.Command definition
        if grep -q '&cobra\.Command{' "$file" 2>/dev/null; then
            # Check if it has Example field
            if ! grep -q 'Example:' "$file"; then
                error "$file: Command missing Example field"
            fi
        fi
    done < <(find internal/cmd -name '*.go' -type f 2>/dev/null)
fi

# ============================================================================
# CODEBASE-WIDE RULES (all internal/ .go files)
# ============================================================================

section "Codebase-Wide Rules (internal/)"

# Rule 4: No iostreams.System()
info "Checking for iostreams.System()..."
# Exclude: factory.go (provides it), test files (legitimate use)
check_pattern 'iostreams\.System()' "Use f.IOStreams() instead of iostreams.System()" "all" "error" "_test.go\|factory.go"

# Rule 5: No shelly.NewService() direct instantiation
info "Checking for shelly.NewService()..."
# Exclude: factory.go (provides it), shelly.go (defines it), test files, completion/ (needs direct access)
check_pattern 'shelly\.NewService()' "Use f.ShellyService() instead of shelly.NewService()" "all" "error" "_test.go\|factory.go\|shelly/shelly.go\|completion/"

# Rule 6: No exec.Command("open"
info "Checking for exec.Command browser calls..."
check_pattern 'exec\.Command("open"' "Use browser.New().Browse() instead of exec.Command" "all" "error"
check_pattern 'exec\.Command("xdg-open"' "Use browser.New().Browse() instead of exec.Command" "all" "error"

# Rule 7: No context.Background() in commands
info "Checking for context.Background()..."
# Exclude: root.go (signal handling), test files, prometheus.go (shutdown handler), install.go (utility function), webhook/server (shutdown handler)
check_pattern 'context\.Background()' "Use cmd.Context() instead of context.Background()" "cmd" "error" "_test.go\|root.go\|prometheus.go\|install.go\|webhook/server"

# Rule 8: No fmt.Println in commands
info "Checking for fmt.Println..."
# Exclude: extension/create and plugin/create (contain template code for generated files)
check_pattern 'fmt\.Println' "Use ios.Print() or ios.Success() instead of fmt.Println" "cmd" "error" "extension/create\|plugin/create"

# Rule 9: No spinner.New() - use iostreams.NewSpinner
info "Checking for spinner.New()..."
# Exclude: progress.go (this is the wrapper itself)
check_pattern 'spinner\.New(' "Use iostreams.NewSpinner() instead of spinner.New()" "all" "error" "progress.go"

# Rule 10: No error suppression with _ = err
info "Checking for error suppression..."
# Match "_ = err" but not blank identifiers in function signatures "_ context" or struct tags
# Pattern: _ = err as a statement (with optional whitespace)
if [[ -d "internal" ]]; then
    results=$(grep -rn --include='*.go' '^\s*_ = err\b\|[^_]\s_ = err\b' internal 2>/dev/null || true)
    if [[ -n "$results" ]]; then
        while IFS= read -r line; do
            # Skip lines that are clearly not error suppression (struct tags, blank params)
            if echo "$line" | grep -qv 'json:\|yaml:\|func.*_ [a-zA-Z]'; then
                error "$line - Handle errors properly, use ios.DebugErr() for non-critical errors"
            fi
        done <<< "$results"
    fi
fi

# Rule 11: No //nolint:errcheck directives (error suppression requires approval)
info "Checking for //nolint:errcheck directives..."
check_pattern '//nolint:errcheck' "nolint:errcheck requires explicit approval - handle errors properly" "all" "warning"

# Rule 12: No TODO comments
info "Checking for TODO comments..."
check_pattern '// TODO' "Fix TODO items, don't defer them" "all" "error"
check_pattern '//TODO' "Fix TODO items, don't defer them" "all" "error"

# Rule 13: No "deferred" or "future work" comments
info "Checking for deferred/future language..."
check_pattern_nocase 'deferred' "Don't defer work - fix issues now" "all" "error"
# Only flag "future" when it indicates deferred work, not legitimate descriptions
check_pattern_nocase 'in the future' "Don't defer work to 'future' - fix issues now" "all" "error"
check_pattern_nocase 'future version' "Don't promise future versions - implement now or remove" "all" "error"
check_pattern_nocase 'future release' "Don't promise future releases - implement now or remove" "all" "error"

# Rule 14: No "best effort" (any case)
info "Checking for 'best effort' language..."
check_pattern_nocase 'best effort' "'best effort' is not acceptable - implement properly" "all" "error"

# ============================================================================
# SUMMARY
# ============================================================================

section "Summary"

if [[ $ERRORS -eq 0 ]] && [[ $WARNINGS -eq 0 ]]; then
    echo -e "${GREEN}All compliance checks passed!${NC}"
    exit 0
fi

if [[ $WARNINGS -gt 0 ]]; then
    echo -e "${YELLOW}Warnings: $WARNINGS${NC}"
fi

if [[ $ERRORS -gt 0 ]]; then
    echo -e "${RED}Errors: $ERRORS${NC}"
    echo ""
    echo "Please fix all errors before proceeding."
    exit 1
fi

exit 0
