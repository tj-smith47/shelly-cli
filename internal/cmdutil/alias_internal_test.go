package cmdutil

import (
	"strings"
	"testing"
)

func TestExpandAliasArgs_AllArgsPlaceholderForwardsEveryArgument(t *testing.T) {
	t.Parallel()
	got, err := expandAliasArgs("device remove $@", []string{"kitchen", "--yes"})
	if err != nil {
		t.Fatalf("expandAliasArgs: %v", err)
	}
	want := []string{"device", "remove", "kitchen", "--yes"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("expandAliasArgs = %v, want %v", got, want)
	}
}

func TestExpandAliasArgs_PositionalThenRest(t *testing.T) {
	t.Parallel()
	got, err := expandAliasArgs("schedule create --timespec \"$1\" $2", []string{"0 7 * * *", "kitchen", "--yes"})
	if err != nil {
		t.Fatalf("expandAliasArgs: %v", err)
	}
	want := []string{"schedule", "create", "--timespec", "0 7 * * *", "kitchen", "--yes"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("expandAliasArgs = %v, want %v", got, want)
	}
}
