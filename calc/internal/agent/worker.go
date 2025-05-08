package agent

import (
	"context"
	"time"

	"github.com/TimofeySar/ya_go_calculate.go/internal/orchestrator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Run(power int, conn *grpc.ClientConn) {
	for i := 0; i < power; i++ {
		go worker(conn)
	}
	select {}
}

func worker(conn *grpc.ClientConn) {
	client := orchestrator.NewTaskServiceClient(conn)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		task, err := client.GetTask(ctx, &orchestrator.Empty{})
		cancel()
		if err != nil {
			if status.Code(err) == codes.NotFound {
				time.Sleep(1 * time.Second)
				continue
			}
			time.Sleep(1 * time.Second)
			continue
		}

		time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
		var result float64
		if task.Arg1 != 0 {
			switch task.Operation {
			case "+":
				result = task.Arg1 + task.Arg2
			case "-":
				result = task.Arg1 - task.Arg2
			case "*":
				result = task.Arg1 * task.Arg2
			case "/":
				if task.Arg2 == 0 {
					continue
				}
				result = task.Arg1 / task.Arg2
			}
		}

		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		_, err = client.SendResult(ctx, &orchestrator.Result{Id: task.Id, Result: result})
		cancel()
		if err != nil {
			continue
		}
	}
}
