// Package output provides output formatting utilities for the CLI.
package output

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

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
func Input(message string, defaultValue string) (string, error) {
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
func MultiSelect(message string, options []string, defaults []string) ([]string, error) {
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
func Editor(message string, defaultValue string) (string, error) {
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

// AskQuestions asks multiple questions at once.
type Question struct {
	Name     string
	Prompt   survey.Prompt
	Validate survey.Validator
}

// Ask asks multiple questions and returns the answers in a map.
func Ask(questions []Question) (map[string]any, error) {
	answers := make(map[string]any)

	var qs []*survey.Question
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

// DeviceSelection helpers

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
