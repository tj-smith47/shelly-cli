## Summary

<!-- Brief description of changes (1-2 sentences) -->

## Type of Change

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to change)
- [ ] Documentation update
- [ ] Refactoring (no functional changes)
- [ ] Test improvement

## Related Issues

<!-- Link to related issues: Fixes #123, Closes #456 -->

## Changes Made

<!-- Bullet points of what was changed -->

-
-
-

## Checklist

### Code Quality

- [ ] I have read [RULES.md](RULES.md) and followed all mandatory rules
- [ ] Factory pattern used (no direct instantiation of IOStreams, ShellyService, Browser)
- [ ] All new commands have `Aliases` and `Example` fields
- [ ] Context comes from `cmd.Context()`, not `context.Background()`
- [ ] All output uses IOStreams methods (no `fmt.Println` in commands)
- [ ] Errors are handled properly (no `_ = err` or `//nolint:errcheck`)
- [ ] Import ordering follows gci format (stdlib, third-party, internal)

### Testing

- [ ] Tests added/updated for new functionality
- [ ] All tests pass: `go test ./...`
- [ ] Linter passes: `golangci-lint run ./...`
- [ ] Manual testing performed (if user-facing)

### Documentation

- [ ] Help text updated for new/modified commands
- [ ] PLAN.md updated (if completing a planned task)
- [ ] CHANGELOG.md updated (if user-facing change)

## Testing Done

<!-- Describe how you tested these changes -->

```bash
# Commands run to test
```

## Screenshots (if applicable)

<!-- For TUI changes or output format changes -->

## Additional Notes

<!-- Any additional context or notes for reviewers -->
