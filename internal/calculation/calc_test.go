package calculation_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/TimofeySar/ya_go_calculate.go/internal/calculation"
)

func TestMain(m *testing.M) {
	os.Setenv("TIME_ADDITION_MS", "1000")
	os.Setenv("TIME_SUBTRACTION_MS", "1000")
	os.Setenv("TIME_MULTIPLICATIONS_MS", "2000")
	os.Setenv("TIME_DIVISIONS_MS", "2000")
	os.Exit(m.Run())
}

func TestInfixToPostfix(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expected    []string
		expectedErr string
	}{
		{"simple addition", "1+1", []string{"1", "1", "+"}, ""},
		{"multiplication priority", "2+2*2", []string{"2", "2", "2", "*", "+"}, ""},
		{"parentheses priority", "(2+2)*2", []string{"2", "2", "+", "2", "*"}, ""},
		{"division", "1/2", []string{"1", "2", "/"}, ""},
		{"complex expression", "(3+5)*(2-1)", []string{"3", "5", "+", "2", "1", "-", "*"}, ""},
		{"empty expression", "", nil, "пустое выражение"},
		{"invalid operator at end", "1+1*", nil, "некорректное выражение: недостаточно операндов"},
		{"double operator", "2+2**2", nil, "некорректный оператор: *"},
		{"unmatched parentheses", "((2+2)", nil, "некорректное выражение: несогласованные скобки"},
		{"invalid symbol", "2+a", nil, "некорректный символ: a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculation.InfixToPostfix(tt.expression)
			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.expectedErr)
				}
				if err.Error() != tt.expectedErr {
					t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestGenerateTasks(t *testing.T) {
	tests := []struct {
		name        string
		postfix     []string
		expected    int
		expectedErr string
	}{
		{"simple addition", []string{"1", "1", "+"}, 1, ""},
		{"multiplication priority", []string{"2", "2", "2", "*", "+"}, 2, ""},
		{"parentheses priority", []string{"2", "2", "+", "2", "*"}, 2, ""},
		{"division", []string{"1", "2", "/"}, 1, ""},
		{"complex expression", []string{"3", "5", "+", "2", "1", "-", "*"}, 3, ""},
		{"insufficient operands", []string{"1", "+"}, 0, "некорректное выражение: недостаточно операндов"},
		{"extra operands", []string{"1", "2", "3", "+"}, 0, "некорректное выражение: лишние операнды"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := calculation.GenerateTasks(tt.postfix)
			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.expectedErr)
				}
				if err.Error() != tt.expectedErr {
					t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(tasks) != tt.expected {
				t.Errorf("expected %d tasks, got %d", tt.expected, len(tasks))
			}
			for _, task := range tasks {
				expectedTime := map[string]int{
					"+": 1000,
					"-": 1000,
					"*": 2000,
					"/": 2000,
				}[task.Operation]
				if task.OperationTime != expectedTime {
					t.Errorf("for operation %q, expected time %d, got %d", task.Operation, expectedTime, task.OperationTime)
				}
			}
		})
	}
}
