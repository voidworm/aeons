package prompt

import "github.com/manifoldco/promptui"

func Prompt(label string, options []string) (int, error) {

	prompt := promptui.Select{
		Label: label,
		Items: options,
	}
	position, _, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	return position, err
}

func PromptCancellable(label string, options []string) (bool, int, error) {

	options = append(options, "Cancel")

	prompt := promptui.Select{
		Label: label,
		Items: options,
	}
	position, selected, err := prompt.Run()
	if err != nil {
		return false, 0, err
	}
	if selected == "Cancel" {
		return true, position, nil
	}

	return false, position, err
}
