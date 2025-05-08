package orchestrator

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/TimofeySar/ya_go_calculate.go/internal/calculation"
)

type Expression struct {
	ID        string
	UserID    string
	Expr      string
	Status    string
	Result    float64
	Tasks     map[string]*calculation.Task
	TaskOrder []string
	Results   []float64
	Postfix   []string
}

func NewExpression(id, userID, expr string) *Expression {
	return &Expression{
		ID:        id,
		UserID:    userID,
		Expr:      expr,
		Status:    "pending",
		Tasks:     make(map[string]*calculation.Task),
		TaskOrder: []string{},
		Results:   []float64{},
		Postfix:   []string{},
	}
}

func (s *Expression) Start(tasksChan chan<- *calculation.Task) {
	postfix, err := calculation.InfixToPostfix(s.Expr)
	if err != nil {
		s.Status = "error"
		return
	}
	s.Postfix = postfix

	tasks, err := calculation.GenerateTasks(postfix)
	if err != nil {
		s.Status = "error"
		return
	}

	for i, task := range tasks {
		fmt.Printf("Task %d: Operation=%s, Arg1=%f, Arg2=%f\n", i+1, task.Operation, task.Arg1, task.Arg2)
		s.Tasks[task.ID] = task
		s.TaskOrder = append(s.TaskOrder, task.ID)
		tasksChan <- task
	}
}

func (e *Expression) UpdateTaskResult(taskID string, result float64) bool {
	if task, exists := e.Tasks[taskID]; exists {
		delete(e.Tasks, taskID)
		if task.Arg1 == 0 && len(e.Results) > 0 {
			intermediateResult := e.Results[len(e.Results)-1]
			switch task.Operation {
			case "*":
				result = intermediateResult * task.Arg2
			case "/":
				result = intermediateResult / task.Arg2
			case "+":
				result = intermediateResult + task.Arg2
			case "-":
				result = intermediateResult - task.Arg2
			}
		}
		e.Results = append(e.Results, result)
		fmt.Printf("Updated task %s with result %f, e.Results=%v\n", taskID, result, e.Results)

		if len(e.Tasks) == 0 {
			e.Status = "completed"
			if len(e.Results) > 0 {
				e.Result = e.Results[len(e.Results)-1]
				fmt.Printf("Final e.Result=%f\n", e.Result)
			}
		}
		return true
	}
	fmt.Printf("Task %s not found\n", taskID)
	return false
}

func generateID() string {
	return "expr-" + uuid.New().String()
}
