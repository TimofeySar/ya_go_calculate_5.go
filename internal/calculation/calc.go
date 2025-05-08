package calculation

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1,omitempty"`
	Arg2          float64 `json:"arg2,omitempty"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

var operationTimes = LoadEnv()

func LoadEnv() map[string]int {
	envVars := map[string]string{
		"TIME_ADDITION_MS":        "1000",
		"TIME_SUBTRACTION_MS":     "1000",
		"TIME_MULTIPLICATIONS_MS": "2000",
		"TIME_DIVISIONS_MS":       "2000",
	}
	times := make(map[string]int)
	for key, defaultVal := range envVars {
		val := os.Getenv(key)
		if val == "" {
			val = defaultVal
		}
		if ms, err := strconv.Atoi(val); err == nil {
			switch key {
			case "TIME_ADDITION_MS":
				times["+"] = ms
			case "TIME_SUBTRACTION_MS":
				times["-"] = ms
			case "TIME_MULTIPLICATIONS_MS":
				times["*"] = ms
			case "TIME_DIVISIONS_MS":
				times["/"] = ms
			}
		}
	}
	return times
}

func InfixToPostfix(expression string) ([]string, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	if expression == "" {
		return nil, errors.New("пустое выражение")
	}

	var postfix []string
	var stack []rune
	precedence := map[rune]int{
		'+': 1,
		'-': 1,
		'*': 2,
		'/': 2,
	}

	isOperator := func(ch rune) bool {
		_, exists := precedence[ch]
		return exists
	}

	openParentheses := 0
	lastChar := ' '

	for i, ch := range expression {
		switch {
		case ch >= '0' && ch <= '9' || ch == '.':
			if lastChar >= '0' && lastChar <= '9' || lastChar == '.' {
				postfix[len(postfix)-1] += string(ch)
			} else {
				postfix = append(postfix, string(ch))
			}
			lastChar = ch
		case isOperator(ch):
			if isOperator(lastChar) || lastChar == '(' || i == 0 {
				return nil, fmt.Errorf("некорректный оператор: %c", ch)
			}
			for len(stack) > 0 && isOperator(stack[len(stack)-1]) && precedence[stack[len(stack)-1]] >= precedence[ch] {
				postfix = append(postfix, string(stack[len(stack)-1]))
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, ch)
			lastChar = ch
		case ch == '(':
			stack = append(stack, ch)
			openParentheses++
			lastChar = ch
		case ch == ')':
			if openParentheses == 0 {
				return nil, errors.New("некорректное выражение: несогласованные скобки")
			}
			for len(stack) > 0 && stack[len(stack)-1] != '(' {
				postfix = append(postfix, string(stack[len(stack)-1]))
				stack = stack[:len(stack)-1]
			}
			stack = stack[:len(stack)-1]
			openParentheses--
			lastChar = ch
		default:
			return nil, fmt.Errorf("некорректный символ: %c", ch)
		}
	}

	if isOperator(lastChar) {
		return nil, errors.New("некорректное выражение: недостаточно операндов")
	}

	if openParentheses > 0 {
		return nil, errors.New("некорректное выражение: несогласованные скобки")
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == '(' {
			return nil, errors.New("некорректное выражение: несогласованные скобки")
		}
		postfix = append(postfix, string(stack[len(stack)-1]))
		stack = stack[:len(stack)-1]
	}

	return postfix, nil
}

func GenerateTasks(postfix []string) ([]*Task, error) {
	var stack []float64
	var tasks []*Task
	taskCounter := 0

	for _, token := range postfix {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, num)
		} else {
			if len(stack) < 2 {
				return nil, errors.New("некорректное выражение: недостаточно операндов")
			}
			arg2 := stack[len(stack)-1]
			arg1 := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			taskCounter++
			task := &Task{
				ID:            fmt.Sprintf("task-%d", taskCounter),
				Arg1:          arg1,
				Arg2:          arg2,
				Operation:     token,
				OperationTime: operationTimes[token],
			}
			tasks = append(tasks, task)
			stack = append(stack, 0)
		}
	}

	if len(stack) != 1 {
		return nil, errors.New("некорректное выражение: лишние операнды")
	}

	return tasks, nil
}
