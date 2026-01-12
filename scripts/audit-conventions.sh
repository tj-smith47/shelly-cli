#!/usr/bin/env bash
# Audit script for shelly-cli convention/rule/pattern compliance
# Based on project rules in docs/development.md and docs/architecture.md
#
# Usage: ./scripts/audit-conventions.sh [path]
#   path: Optional directory to audit (default: entire repo)
#
# This script should be run as the final verification step.
# Any error means there is a compliance issue that MUST be fixed.
# REMINDER: No issue is pre-existing - all code in this repo was written by Claude.

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default to repo root if no path specified
AUDIT_PATH="${1:-.}"
AUDIT_ONLY="${AUDIT_ONLY:-false}"
ERRORS=0
WARNINGS=0

# ==============================================================================
# HELPER FUNCTIONS
# ==============================================================================

error() {
    echo -e "${RED}ERROR:${NC} $1"
    ((ERRORS++)) || true
}

warn() {
    echo -e "${YELLOW}WARNING:${NC} $1"
    ((WARNINGS++)) || true
}

success() {
    echo -e "${GREEN}OK:${NC} $1"
}

section() {
    echo ""
    echo -e "${BLUE}--- $1 ---${NC}"
}

# Search Go files, excluding tests and vendor
# Usage: search_go [-i] <pattern> [path]
#   -i: case-insensitive matching
search_go() {
    local case_flag=""
    if [[ "$1" == "-i" ]]; then
        case_flag="-i"
        shift
    fi
    local pattern="$1"
    local path="${2:-internal/}"
    grep -rn $case_flag "$pattern" "$path" 2>/dev/null | grep "\.go:" | grep -v "_test\.go" | grep -v "vendor/" || true
}

# Search only cmd/ Go files, excluding tests
search_cmd() {
    local pattern="$1"
    grep -rn "$pattern" internal/cmd/ 2>/dev/null | grep "\.go:" | grep -v "_test\.go" || true
}

# Count matches
count_matches() {
    local result="$1"
    if [[ -z "$result" ]]; then
        echo "0"
    else
        echo "$result" | wc -l | tr -d ' '
    fi
}

# Display results with limit
show_results() {
    local result="$1"
    local limit="${2:-10}"
    if [[ -n "$result" ]]; then
        echo "$result" | head -"$limit"
        local total
        total=$(count_matches "$result")
        if [[ "$total" -gt "$limit" ]]; then
            echo "  ... and $((total - limit)) more"
        fi
    fi
}

# ==============================================================================
# MAIN SCRIPT
# ==============================================================================

echo "=========================================="
echo "  SHELLY-CLI CONVENTION AUDIT"
echo "=========================================="
echo ""
echo "Auditing: $AUDIT_PATH"

# ==============================================================================
# SECTION 1: Stub/Incomplete Code Checks
# ==============================================================================
section "Stub/Incomplete Code Checks"

# Check for TODO comments
TODOS=$(search_go "TODO" "internal/")
if [[ -n "$TODOS" ]]; then
    error "Found TODO comments that need to be addressed:"
    show_results "$TODOS"
else
    success "No TODO comments found"
fi

# Check for FIXME comments
FIXMES=$(search_go "FIXME" "internal/")
if [[ -n "$FIXMES" ]]; then
    error "Found FIXME comments that need to be addressed:"
    show_results "$FIXMES"
else
    success "No FIXME comments found"
fi

# Check for stub implementations
STUBS=$(search_go "stub\|// stub\|Stub\(" "internal/")
if [[ -n "$STUBS" ]]; then
    error "Found stub implementations:"
    show_results "$STUBS"
else
    success "No stub implementations found"
fi

# Check for placeholder comments (exclude form UI placeholders, term.go helper, and legitimate placeholder patterns)
PLACEHOLDERS=$(search_go "placeholder\|PLACEHOLDER" "internal/" | \
    grep -v "internal/tui/components/form/" | \
    grep -v "internal/term/term.go" | \
    grep -v "argument placeholders" | \
    grep -v "SubstituteVariables" | \
    grep -v "LabelPlaceholder" | \
    grep -v "ZeroAsPlaceholder" | \
    grep -v "getPlaceholder" | \
    grep -vi "show placeholder" | \
    grep -v "FormatPlaceholder" | \
    grep -v "Placeholder.*string" | \
    grep -v "as placeholder" | \
    grep -v "\$N placeholder" | \
    grep -v "placeholder :=" | \
    grep -v "if placeholder ==" | \
    grep -v "return placeholder" | \
    grep -v "Contains.*placeholder" | \
    grep -v "ReplaceAll.*placeholder" || true)
if [[ -n "$PLACEHOLDERS" ]]; then
    error "Found placeholder code:"
    show_results "$PLACEHOLDERS"
else
    success "No placeholder code found"
fi

# Check for deferred work (exclude legitimate patterns)
DEFERRED=$(search_go "future\|deferred\|later\|eventually\|nice to have\|optional.*implement" "internal/" | \
    grep -v "later with:" | \
    grep -v "deferredEvent" | \
    grep -v "deferred event" | \
    grep -v "for future use" | \
    grep -v "for future calls" | \
    grep -v "# Re-enable" | \
    grep -v "restore later" | \
    grep -v "manifest.go.*optional" || true)
if [[ -n "$DEFERRED" ]]; then
    error "Found deferred/future work references:"
    show_results "$DEFERRED"
else
    success "No deferred/future work references found"
fi

# Check for incomplete implementation mentions (exclude tracked TUI work in TUI-PLAN.md)
INCOMPLETE=$(search_go "full implementation\|complete implementation\|proper implementation\|real implementation" "internal/" | \
    grep -v "internal/tui/components/modal/model.go" || true)
if [[ -n "$INCOMPLETE" ]]; then
    error "Found mentions of incomplete implementation:"
    show_results "$INCOMPLETE"
else
    success "No incomplete implementation mentions found"
fi

# Check for legacy/backward compatibility/deprecated patterns (should be cleaned up, not deferred)
LEGACY=$(search_go -i "legacy\|backward.*compat\|deprecated\|kept for compat" "internal/" | \
    grep -vi "IsLegacy\|legacyHost\|legacyDevice" || true)
if [[ -n "$LEGACY" ]]; then
    error "Found legacy/deprecated patterns (clean up, don't defer):"
    show_results "$LEGACY"
else
    success "No legacy/deprecated patterns found"
fi

# Check for panic placeholders
PANICS=$(search_go 'panic\(".*not implemented' "internal/")
if [[ -n "$PANICS" ]]; then
    error "Found panic placeholders for unimplemented code:"
    show_results "$PANICS"
else
    success "No panic placeholders found"
fi

# ==============================================================================
# SECTION 2: Factory Pattern Checks
# ==============================================================================
section "Factory Pattern Checks"

# Check for NewCmd naming (should be NewCommand)
BAD_NAMING=$(search_cmd "func NewCmd[A-Z]")
if [[ -n "$BAD_NAMING" ]]; then
    error "Found 'NewCmdXxx' naming instead of 'NewCommand':"
    show_results "$BAD_NAMING"
else
    success "All command constructors use 'NewCommand' naming"
fi

# Check for run() with factory as separate parameter (old pattern)
OLD_RUN_PATTERN=$(search_cmd "func run(ctx context.Context, f \*cmdutil\.Factory")
if [[ -n "$OLD_RUN_PATTERN" ]]; then
    error "Found old run(ctx, f, ...) pattern instead of run(ctx, opts):"
    show_results "$OLD_RUN_PATTERN"
else
    success "All run() functions use Factory-in-Options pattern"
fi

# ==============================================================================
# SECTION 3: IOStreams Usage Checks
# ==============================================================================
section "IOStreams Usage Checks"

# Check for fmt.Println anywhere in internal/ (exclude scaffold.go template code)
FMT_PRINTLN=$(search_go "fmt\.Println" "internal/" | \
    grep -v "internal/plugins/scaffold/scaffold.go" || true)
if [[ -n "$FMT_PRINTLN" ]]; then
    error "Found fmt.Println (should use ios methods):"
    show_results "$FMT_PRINTLN"
else
    success "No fmt.Println in internal/"
fi

# Check for fmt.Printf to stdout (exclude Sprintf/Fprintf and scaffold.go template code)
FMT_PRINTF=$(search_go "fmt\.Printf" "internal/" | \
    grep -v "fmt\.Fprintf" | \
    grep -v "fmt\.Sprintf" | \
    grep -v "internal/plugins/scaffold/scaffold.go" || true)
if [[ -n "$FMT_PRINTF" ]]; then
    error "Found fmt.Printf (should use ios methods):"
    show_results "$FMT_PRINTF"
else
    success "No direct fmt.Printf in internal/"
fi

# Check for iostreams.System() direct instantiation (exclude factory.go where it's created)
DIRECT_IOSTREAMS=$(search_go "iostreams\.System()" "internal/" | \
    grep -v "internal/cmdutil/factory.go" || true)
if [[ -n "$DIRECT_IOSTREAMS" ]]; then
    error "Found iostreams.System() direct instantiation:"
    show_results "$DIRECT_IOSTREAMS"
else
    success "No direct iostreams.System() instantiation"
fi

# Check for shelly.NewService() direct instantiation (exclude factory.go and completion.go)
DIRECT_SERVICE=$(search_go "shelly\.NewService()" "internal/" | \
    grep -v "internal/cmdutil/factory.go" | \
    grep -v "internal/completion/completion.go" || true)
if [[ -n "$DIRECT_SERVICE" ]]; then
    error "Found shelly.NewService() direct instantiation:"
    show_results "$DIRECT_SERVICE"
else
    success "No direct shelly.NewService() instantiation"
fi

# ==============================================================================
# SECTION 4: Context Usage Checks
# ==============================================================================
section "Context Usage Checks"

# Check for context.Background() in commands (exclude legitimate uses)
# Legitimate: plugins (run outside cmd), completion (outside cmd), root.go (context origin),
# automation (background event listeners), telemetry (background send), shutdown contexts
CTX_BACKGROUND=$(search_go "context\.Background()" "internal/" | \
    grep -v "internal/plugins/" | \
    grep -v "internal/completion/" | \
    grep -v "internal/cmd/root.go" | \
    grep -v "internal/shelly/automation/" | \
    grep -v "internal/telemetry/" | \
    grep -vi "shutdown" || true)
if [[ -n "$CTX_BACKGROUND" ]]; then
    error "Found context.Background() (should use cmd.Context()):"
    show_results "$CTX_BACKGROUND"
else
    success "No context.Background() in command code"
fi

# ==============================================================================
# SECTION 5: Architecture Separation of Concerns
# ==============================================================================
section "Architecture Separation of Concerns"

# cmd/ should only have NewCommand and run functions
# Exclude root.go (special CLI entry point)
OTHER_FUNCS=$(grep -rn "^func [A-Za-z]" internal/cmd/ 2>/dev/null | grep "\.go:" | grep -v "_test\.go" | grep -v "func NewCommand" | grep -v "func run(" | grep -v "func init(" | grep -v "root\.go" || true)
if [[ -n "$OTHER_FUNCS" ]]; then
    error "Found functions other than NewCommand/run in cmd/ (move to appropriate package per architecture.md):"
    show_results "$OTHER_FUNCS"
else
    success "cmd/ contains only NewCommand and run functions"
fi

# Check for display functions anywhere outside term/
DISPLAY_OUTSIDE_TERM=$(search_go "func [Dd]isplay" "internal/" | grep -v "internal/term" | grep -v "internal/tui" || true)
if [[ -n "$DISPLAY_OUTSIDE_TERM" ]]; then
    error "Found Display functions outside term/ (should be in term/):"
    show_results "$DISPLAY_OUTSIDE_TERM"
else
    success "Display functions correctly placed in term/"
fi

# Check for format functions in cmd/ (should be in output/)
FORMAT_IN_CMD=$(search_cmd "func [Ff]ormat")
if [[ -n "$FORMAT_IN_CMD" ]]; then
    error "Found Format functions in cmd/ (should be in output/):"
    show_results "$FORMAT_IN_CMD"
else
    success "No Format functions in cmd/"
fi

# Check for HTTP/client logic in cmd/ (should be in shelly/ or client/)
HTTP_IN_CMD=$(search_cmd "http\.Get\|http\.Post\|http\.Client\|http\.NewRequest")
if [[ -n "$HTTP_IN_CMD" ]]; then
    error "Found HTTP logic in cmd/ (should be in shelly/ or client/):"
    show_results "$HTTP_IN_CMD"
else
    success "No HTTP logic in cmd/"
fi

# Check for direct exec.Command outside browser/ (use browser package)
EXEC_CMD=$(search_go 'exec\.Command\(' "internal/" | grep -v "internal/browser" | grep -v "internal/plugins" || true)
if [[ -n "$EXEC_CMD" ]]; then
    warn "Found exec.Command outside browser/plugins (consider browser package):"
    show_results "$EXEC_CMD" 5
fi

# ==============================================================================
# SECTION 6: Error Handling Checks
# ==============================================================================
section "Error Handling Checks"

# Check for //nolint:errcheck without approval
NOLINT_ERRCHECK=$(search_go "//nolint:errcheck" "internal/")
if [[ -n "$NOLINT_ERRCHECK" ]]; then
    warn "Found //nolint:errcheck (requires approval):"
    show_results "$NOLINT_ERRCHECK" 5
fi

# Check for _ = err pattern
SUPPRESSED_ERR=$(search_go "_ = err" "internal/")
if [[ -n "$SUPPRESSED_ERR" ]]; then
    error "Found '_ = err' error suppression (use ios.DebugErr()):"
    show_results "$SUPPRESSED_ERR"
else
    success "No '_ = err' error suppression"
fi

# ==============================================================================
# SECTION 7: Command Requirements Checks
# ==============================================================================
section "Command Requirements Checks"

# Check for commands missing Aliases in leaf command files
CMD_FILES=$(find internal/cmd -maxdepth 3 -name "*.go" ! -name "*_test.go" ! -name "root.go" -type f 2>/dev/null || true)
MISSING_ALIASES=""
for f in $CMD_FILES; do
    if grep -q "cobra\.Command{" "$f" && ! grep -q "Aliases:" "$f"; then
        MISSING_ALIASES="${MISSING_ALIASES}${f}\n"
    fi
done
if [[ -n "$MISSING_ALIASES" ]]; then
    warn "Commands potentially missing Aliases:"
    echo -e "$MISSING_ALIASES" | head -10
else
    success "All commands have Aliases"
fi

# Check for commands missing Example
MISSING_EXAMPLES=""
for f in $CMD_FILES; do
    if grep -q "cobra\.Command{" "$f" && ! grep -q "Example:" "$f"; then
        MISSING_EXAMPLES="${MISSING_EXAMPLES}${f}\n"
    fi
done
if [[ -n "$MISSING_EXAMPLES" ]]; then
    warn "Commands potentially missing Examples:"
    echo -e "$MISSING_EXAMPLES" | head -10
else
    success "All commands have Examples"
fi

# ==============================================================================
# SECTION 8: Test Isolation Checks
# ==============================================================================
section "Test Isolation Checks"

# Search test files only
search_test() {
    local pattern="$1"
    local path="${2:-internal/}"
    grep -rn "$pattern" "$path" 2>/dev/null | grep "_test\.go:" || true
}

# Check for t.TempDir() usage (should use afero instead)
# Exclude: migrate/validate/validate_test.go:508 - TestRun_PermissionDenied requires real fs for permission testing
TEMP_DIR=$(search_test "t\.TempDir()" "internal/" | \
    grep -v "internal/cmd/migrate/validate/validate_test.go" || true)
TEMP_DIR_COUNT=$(count_matches "$TEMP_DIR")
if [[ "$TEMP_DIR_COUNT" -gt 0 ]]; then
    warn "Found $TEMP_DIR_COUNT uses of t.TempDir() (prefer afero.NewMemMapFs with SetFs):"
    show_results "$TEMP_DIR" 5
else
    success "No t.TempDir() usage (using afero for test isolation)"
fi

# Check for os.MkdirTemp usage in tests
# Legitimate exclusions (execute actual scripts/binaries, need real filesystem):
# - plugins/hooks_test.go, plugins/executor_test.go: execute hook scripts
# - shelly/dispatch_test.go: dispatches to plugin processes
MKDIR_TEMP=$(search_test "os\.MkdirTemp" "internal/" | \
    grep -v "internal/plugins/hooks_test.go" | \
    grep -v "internal/plugins/executor_test.go" | \
    grep -v "internal/shelly/dispatch_test.go" || true)
MKDIR_TEMP_COUNT=$(count_matches "$MKDIR_TEMP")
if [[ "$MKDIR_TEMP_COUNT" -gt 0 ]]; then
    warn "Found $MKDIR_TEMP_COUNT uses of os.MkdirTemp in tests (prefer afero):"
    show_results "$MKDIR_TEMP" 5
else
    success "No os.MkdirTemp usage (using afero for test isolation)"
fi

# Check for os.WriteFile usage in tests (should use afero.WriteFile)
# Legitimate exclusions (create executable scripts for process execution):
# - plugins/hooks_test.go, plugins/executor_test.go: create executable hook scripts
# - shelly/dispatch_test.go: creates plugin executables
# - migrate/validate: permission testing on real filesystem
OS_WRITEFILE=$(search_test "os\.WriteFile" "internal/" | \
    grep -v "internal/plugins/hooks_test.go" | \
    grep -v "internal/plugins/executor_test.go" | \
    grep -v "internal/shelly/dispatch_test.go" | \
    grep -v "internal/cmd/migrate/validate/" || true)
OS_WRITEFILE_COUNT=$(count_matches "$OS_WRITEFILE")
if [[ "$OS_WRITEFILE_COUNT" -gt 0 ]]; then
    warn "Found $OS_WRITEFILE_COUNT uses of os.WriteFile in tests (prefer afero.WriteFile with SetFs):"
    show_results "$OS_WRITEFILE" 5
else
    success "No os.WriteFile usage (using afero for test isolation)"
fi

# ==============================================================================
# SECTION 9: Build, Lint, Test, Docs
# ==============================================================================
if [[ $AUDIT_ONLY == "false" ]]; then
    echo ""
    echo "=========================================="
    echo "  BUILD, LINT, TEST"
    echo "=========================================="

    section "Building"
    if go build ./...; then
        success "Build passed"
    else
        error "Build failed"
    fi

    section "Linting"
    if golangci-lint run --timeout 5m ./... 2>&1; then
        success "Lint passed"
    else
        error "Lint failed"
    fi

    section "Testing"
    if go test -race ./... 2>&1; then
        success "Tests passed"
    else
        error "Tests failed"
    fi

    section "Generating Docs"
    if make docs 2>&1; then
        success "Docs generated"
    else
        error "Docs generation failed"
    fi
fi

# ==============================================================================
# SUMMARY
# ==============================================================================
echo ""
echo "=========================================="
echo "  AUDIT SUMMARY"
echo "=========================================="
echo ""

if [[ $ERRORS -gt 0 ]]; then
    echo -e "${RED}FAILED:${NC} $ERRORS error(s), $WARNINGS warning(s)"
    echo ""
    echo -e "${RED}REMINDER:${NC} No issue is pre-existing. All code in this repository"
    echo "was written by Claude. You are responsible for fixing ALL bugs."
    echo ""
    exit 1
elif [[ $WARNINGS -gt 0 ]]; then
    echo -e "${YELLOW}PASSED WITH WARNINGS:${NC} $WARNINGS warning(s)"
    echo "Please review warnings above and fix if appropriate."
    exit 0
else
    echo -e "${GREEN}PASSED:${NC} All convention checks passed!"
    exit 0
fi
