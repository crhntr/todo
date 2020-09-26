// +build js,wasm

package main

import (
	"fmt"
	"html/template"

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

	loadTasks(tmp)

	window.AddEventListenerFunc("click", func(event window.Event) {
		if button := window.Element(event.Target()).Closest(".task-transition"); !button.IsNull() {
			handleTaskTransition(tmp, button)
		}
	})

	window.AddEventListenerFunc("change", func(event window.Event) {
		if checkbox := window.Element(event.Target()).Closest(taskFilterCheckboxSelector); !checkbox.IsNull() {
			handleToggleShowState()
		}
	})

	select {}
}

func handleTaskTransition(tmp *template.Template, button window.Element) {
	taskEl := button.Closest(".task")

	task, err := LoadTask(button.Closest(".task"))
	if err != nil {
		fmt.Println(err)
		return
	}

	stateTransition := button.Attribute("data-state-transition")

	switch stateTransition {
	case "start":
		task.State, err = task.State.Start()
	case "finish":
		task.State, err = task.State.Finish()
	case "review":
		task.State, err = task.State.Review(button.Attribute("data-review") == "pass")
	default:
		return
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	newTaskEl, err := window.Document.NewElementFromTemplate(tmp, "task", task)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !TaskFilter()(task) {
		newTaskEl.Set("style", "display: none;")
	}

	taskEl.Parent().ReplaceChild(newTaskEl, taskEl)

	updateStateCounts()
}

func handleToggleShowState() {
	show := TaskFilter()

	for _, taskEl := range window.Document.QuerySelectorAll(".task") {
		task, err := LoadTask(taskEl)
		if err != nil {
			continue
		}

		if !show(task) {
			taskEl.Set("style", "display: none;")
			continue
		}

		taskEl.Set("style", "display: block;")
	}

	updateStateCounts()
}

const (
	taskFilterCheckboxSelector = `.task-filter>input[type="checkbox"]`
)

func loadTasks(tmp *template.Template) {
	tasks := []todo.Task{
		{
			ID:    primitive.NewObjectID(),
			Title: "Task 1",
			State: todo.TaskStateTODO,
		},
		{
			ID:    primitive.NewObjectID(),
			Title: "Task 2",
			State: todo.TaskStateTODO,
		},
		{
			ID:    primitive.NewObjectID(),
			Title: "Task 3",
			State: todo.TaskStateTODO,
		},
	}

	show := TaskFilter()

	tasksEL := window.Document.QuerySelector(".tasks")
	for _, task := range tasks {
		taskEl, err := window.Document.NewElementFromTemplate(tmp, "task", task)
		if err != nil {
			fmt.Printf("failed to render task %s: %s", task.ID, err)
			continue
		}

		if !show(task) {
			taskEl.Set("style", "display: none;")
		}

		tasksEL.AppendChild(taskEl)
	}

	updateStateCounts()
}

func LoadTask(el window.Element) (todo.Task, error) {
	if el.IsNull() {
		return todo.Task{}, nil
	}

	taskID, err := primitive.ObjectIDFromHex(el.Attribute("data-id"))
	if err != nil {
		return todo.Task{}, err
	}

	task := todo.Task{
		ID:    taskID,
		State: todo.TaskState(el.Attribute("data-state")),
		Title: el.QuerySelector(".title").Get("value").String(),
	}

	return task, nil
}

func TaskFilter() func(task todo.Task) bool {
	checkboxes := window.Document.QuerySelectorAll(taskFilterCheckboxSelector)

	showStates := make(map[todo.TaskState]struct{})

	for _, checkbox := range checkboxes {
		if checkbox.Get("checked").Truthy() {
			showStates[todo.TaskState(checkbox.Attribute("name"))] = struct{}{}
		}
	}

	return func(task todo.Task) bool {
		_, ok := showStates[task.State]
		return ok
	}
}

func updateStateCounts() {
	stateCounts := map[todo.TaskState]int{
		todo.TaskStateTODO:   0,
		todo.TaskStateActive: 0,
		todo.TaskStateReview: 0,
		todo.TaskStateDone:   0,
	}

	for _, taskEl := range window.Document.QuerySelectorAll(".task") {
		task, err := LoadTask(taskEl)
		if err != nil {
			continue
		}

		stateCounts[task.State] = stateCounts[task.State] + 1
	}

	checkboxSelector := `.task-filter>input[type="checkbox"][name=%q]`
	for state, count := range stateCounts {
		label := window.Document.QuerySelector(checkboxSelector+`+label`, state)

		if count == 0 {
			label.Set("style", "display: none;")
			continue
		}

		label.Set("style", "display: block;")

		countEl := label.QuerySelector(".count")
		countEl.SetInnerHTMLf("%d", count)
	}
}
