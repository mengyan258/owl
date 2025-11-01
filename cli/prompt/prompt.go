package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ValidatorFunc 验证函数类型
type ValidatorFunc func(string) error

// Input 获取用户输入
func Input(question, defaultValue string, validator ValidatorFunc) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// 显示问题和默认值
		if defaultValue != "" {
			fmt.Printf("? %s (%s): ", question, defaultValue)
		} else {
			fmt.Printf("? %s: ", question)
		}

		// 读取用户输入
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)

		// 如果用户没有输入，使用默认值
		if input == "" && defaultValue != "" {
			input = defaultValue
		}

		// 验证输入
		if validator != nil {
			if err := validator(input); err != nil {
				fmt.Printf("  ❌ %s\n", err.Error())
				continue
			}
		}

		return input, nil
	}
}

// Confirm 获取用户确认
func Confirm(question string, defaultValue bool) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		defaultStr := "y/N"
		if defaultValue {
			defaultStr = "Y/n"
		}

		fmt.Printf("? %s (%s): ", question, defaultStr)

		input, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		input = strings.TrimSpace(strings.ToLower(input))

		// 如果用户没有输入，使用默认值
		if input == "" {
			return defaultValue, nil
		}

		switch input {
		case "y", "yes", "true", "1":
			return true, nil
		case "n", "no", "false", "0":
			return false, nil
		default:
			fmt.Println("  ❌ Please enter y/yes or n/no")
			continue
		}
	}
}

// Select 选择列表
func Select(question string, options []string, defaultValue string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("? %s\n", question)

		// 显示选项
		for i, option := range options {
			marker := " "
			if strings.HasPrefix(option, defaultValue) {
				marker = "❯"
			}
			fmt.Printf("  %s %d) %s\n", marker, i+1, option)
		}

		fmt.Print("  Enter choice (1-", len(options), "): ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)

		// 如果用户没有输入，查找默认值
		if input == "" {
			for _, option := range options {
				if strings.HasPrefix(option, defaultValue) {
					return option, nil
				}
			}
			// 如果没找到默认值，使用第一个选项
			if len(options) > 0 {
				return options[0], nil
			}
		}

		// 解析用户输入的数字
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(options) {
			fmt.Printf("  ❌ Please enter a number between 1 and %d\n", len(options))
			continue
		}

		return options[choice-1], nil
	}
}

// MultiSelect 多选列表
func MultiSelect(question string, options []string, defaultValues []string) ([]string, error) {
	reader := bufio.NewReader(os.Stdin)
	selected := make(map[int]bool)

	// 设置默认选中项
	for _, defaultVal := range defaultValues {
		for i, option := range options {
			if strings.HasPrefix(option, defaultVal) {
				selected[i] = true
				break
			}
		}
	}

	for {
		fmt.Printf("? %s (use space to select, enter to confirm)\n", question)

		// 显示选项
		for i, option := range options {
			marker := "◯"
			if selected[i] {
				marker = "◉"
			}
			fmt.Printf("  %s %d) %s\n", marker, i+1, option)
		}

		fmt.Print("  Enter choices (space-separated numbers, or 'done'): ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		input = strings.TrimSpace(input)

		if input == "done" || input == "" {
			// 返回选中的选项
			var result []string
			for i, option := range options {
				if selected[i] {
					result = append(result, option)
				}
			}
			return result, nil
		}

		// 解析用户输入的数字
		choices := strings.Fields(input)
		for _, choiceStr := range choices {
			choice, err := strconv.Atoi(choiceStr)
			if err != nil || choice < 1 || choice > len(options) {
				fmt.Printf("  ❌ Invalid choice: %s\n", choiceStr)
				continue
			}

			// 切换选中状态
			index := choice - 1
			selected[index] = !selected[index]
		}
	}
}

// Password 获取密码输入（简化版，实际应该隐藏输入）
func Password(question string) (string, error) {
	fmt.Printf("? %s: ", question)
	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}
