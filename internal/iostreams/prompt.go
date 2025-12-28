// Package iostreams provides unified I/O handling for the CLI.
package iostreams

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Prompt functions for interactive user input.
// These wrap the survey library for consistent prompting.

// Confirm prompts the user for a yes/no confirmation.
func Confirm(message string, defaultValue bool) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
		Default: defaultValue,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

// ConfirmDanger prompts for confirmation on dangerous operations.
// Requires typing "yes" to confirm.
func ConfirmDanger(message string) (bool, error) {
	var result string
	prompt := &survey.Input{
		Message: message + " Type 'yes' to confirm:",
	}
	err := survey.AskOne(prompt, &result)
	if err != nil {
		return false, err
	}
	return result == "yes", nil
}

// Input prompts the user for text input.
func Input(message, defaultValue string) (string, error) {
	var result string
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

// InputRequired prompts the user for required text input.
func InputRequired(message string) (string, error) {
	var result string
	prompt := &survey.Input{
		Message: message,
	}
	err := survey.AskOne(prompt, &result, survey.WithValidator(survey.Required))
	return result, err
}

// Password prompts the user for a password (hidden input).
func Password(message string) (string, error) {
	var result string
	prompt := &survey.Password{
		Message: message,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

// Select prompts the user to select one option from a list.
func Select(message string, options []string, defaultIndex int) (string, error) {
	var result string
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}
	if defaultIndex >= 0 && defaultIndex < len(options) {
		prompt.Default = options[defaultIndex]
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

// SelectIndex prompts the user to select one option and returns the index.
func SelectIndex(message string, options []string, defaultIndex int) (int, error) {
	var result int
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}
	if defaultIndex >= 0 && defaultIndex < len(options) {
		prompt.Default = options[defaultIndex]
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

// MultiSelect prompts the user to select multiple options from a list.
func MultiSelect(message string, options, defaults []string) ([]string, error) {
	var result []string
	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
		Default: defaults,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

// Editor opens the user's editor for multi-line input.
func Editor(message, defaultValue string) (string, error) {
	var result string
	prompt := &survey.Editor{
		Message:       message,
		Default:       defaultValue,
		HideDefault:   true,
		AppendDefault: true,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

// Question represents a prompt question for multiple question surveys.
type Question struct {
	Name     string
	Prompt   survey.Prompt
	Validate survey.Validator
}

// Ask asks multiple questions and returns the answers in a map.
func Ask(questions []Question) (map[string]any, error) {
	answers := make(map[string]any)

	qs := make([]*survey.Question, 0, len(questions))
	for _, q := range questions {
		sq := &survey.Question{
			Name:   q.Name,
			Prompt: q.Prompt,
		}
		if q.Validate != nil {
			sq.Validate = q.Validate
		}
		qs = append(qs, sq)
	}

	err := survey.Ask(qs, &answers)
	return answers, err
}

// Device selection helpers

// SelectDevice prompts the user to select a device from a list.
func SelectDevice(message string, devices []string) (string, error) {
	if len(devices) == 0 {
		return "", fmt.Errorf("no devices available")
	}
	return Select(message, devices, 0)
}

// SelectDevices prompts the user to select multiple devices from a list.
func SelectDevices(message string, devices []string) ([]string, error) {
	if len(devices) == 0 {
		return nil, fmt.Errorf("no devices available")
	}
	return MultiSelect(message, devices, nil)
}

// Credential prompts for username and password.
func Credential(usernameMsg, passwordMsg string) (username, password string, err error) {
	username, err = Input(usernameMsg, "admin")
	if err != nil {
		return "", "", err
	}
	password, err = Password(passwordMsg)
	if err != nil {
		return "", "", err
	}
	return username, password, nil
}

// PromptTypedInput prompts for a value with type parsing (number, boolean, string).
// Used for script template variable configuration.
func PromptTypedInput(message, defaultStr, valueType string) (any, error) {
	input, err := Input(message, defaultStr)
	if err != nil {
		return nil, err
	}

	input = strings.TrimSpace(input)
	if input == "" || input == defaultStr {
		// Return default as-is (caller handles nil case)
		return defaultStr, nil
	}

	// Parse based on type
	switch valueType {
	case "number":
		if strings.Contains(input, ".") {
			return strconv.ParseFloat(input, 64)
		}
		return strconv.Atoi(input)
	case "boolean":
		return strconv.ParseBool(input)
	default:
		return input, nil
	}
}

// IOStreams prompt methods

// Confirm prompts the user for a yes/no confirmation using this IOStreams.
// Returns false if the terminal doesn't support prompts.
func (s *IOStreams) Confirm(message string, defaultValue bool) (bool, error) {
	if !s.CanPrompt() {
		return defaultValue, nil
	}
	return Confirm(message, defaultValue)
}

// ConfirmDanger prompts for confirmation on dangerous operations.
// Requires typing "yes" to confirm.
// Returns false if the terminal doesn't support prompts.
func (s *IOStreams) ConfirmDanger(message string) (bool, error) {
	if !s.CanPrompt() {
		return false, nil
	}
	return ConfirmDanger(message)
}

// Input prompts the user for text input.
// Returns the default value if the terminal doesn't support prompts.
func (s *IOStreams) Input(message, defaultValue string) (string, error) {
	if !s.CanPrompt() {
		return defaultValue, nil
	}
	return Input(message, defaultValue)
}

// Select prompts the user to select one option from a list.
// Returns the default option if the terminal doesn't support prompts.
func (s *IOStreams) Select(message string, options []string, defaultIndex int) (string, error) {
	if !s.CanPrompt() {
		if defaultIndex >= 0 && defaultIndex < len(options) {
			return options[defaultIndex], nil
		}
		if len(options) > 0 {
			return options[0], nil
		}
		return "", fmt.Errorf("no options available")
	}
	return Select(message, options, defaultIndex)
}

// MultiSelect prompts the user to select multiple options from a list.
// Returns the defaults if the terminal doesn't support prompts.
func (s *IOStreams) MultiSelect(message string, options, defaults []string) ([]string, error) {
	if !s.CanPrompt() {
		return defaults, nil
	}
	return MultiSelect(message, options, defaults)
}

// PromptTypedInput prompts for a typed value.
// Returns the default value if the terminal doesn't support prompts.
func (s *IOStreams) PromptTypedInput(message, defaultStr, valueType string) (any, error) {
	if !s.CanPrompt() {
		return defaultStr, nil
	}
	return PromptTypedInput(message, defaultStr, valueType)
}

// PromptScriptVariables prompts for script template variable values.
// Returns a map of variable names to their values, starting with defaults
// and updating with user input when configure is true.
func (s *IOStreams) PromptScriptVariables(variables []config.ScriptVariable, configure bool) map[string]any {
	values := make(map[string]any, len(variables))
	for _, v := range variables {
		values[v.Name] = v.Default
	}

	if !configure || len(variables) == 0 || !s.CanPrompt() {
		return values
	}

	s.Title("Configure Template Variables")
	s.Println()

	for _, v := range variables {
		defaultStr := fmt.Sprintf("%v", v.Default)
		prompt := v.Name
		if v.Description != "" {
			prompt += fmt.Sprintf(" (%s)", v.Description)
		}
		result, err := PromptTypedInput(prompt, defaultStr, v.Type)
		if err != nil {
			continue // Keep default on error
		}
		if resultStr, ok := result.(string); !ok || resultStr != defaultStr {
			values[v.Name] = result
		}
	}
	s.Println()

	return values
}
