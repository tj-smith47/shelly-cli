// Package iostreams provides unified I/O handling for the CLI.
package iostreams

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
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
