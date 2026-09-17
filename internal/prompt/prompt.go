package prompt

import "github.com/manifoldco/promptui"

func PromptUser(label string, options []string, cancelable bool) (int, error) {
	if cancelable {
		options = append(options, "Cancel")
	}
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
