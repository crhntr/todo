// +build js,wasm

package main

import (
	"bytes"
	"fmt"

	"github.com/crhntr/window"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/crhntr/todo"
)

func main() {
	tmp, err := window.LoadTemplates(nil, `script[type="text/go-template"]`)
	if err != nil {
		fmt.Println(err)
		return
	}

	tasks := []todo.Task{
		{
			ID:    primitive.NewObjectID(),
			Title: "Task 1",
		},
		{
			ID:    primitive.NewObjectID(),
			Title: "Task 2",
		},
		{
			ID:    primitive.NewObjectID(),
			Title: "Task 3",
		},
	}

	tasksEL := window.Document.QuerySelector(".tasks")
	var buf bytes.Buffer
	for _, task := range tasks {
		buf.Reset()

		if err := tmp.ExecuteTemplate(&buf, "task", task); err != nil {
			fmt.Printf("failed to render task %s: %s", task.ID, err)
			continue
		}

		tasksEL.AppendHTML(buf.String())
	}
}
