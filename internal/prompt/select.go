package prompt

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

// ShowSelectionPrompt displays a selection prompt with dynamic items and label.
func ShowSelectionPrompt(items []string, label string) (string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", err
	}

	return result, nil
}
