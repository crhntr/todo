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
			updateStateCounts()
		}

		if button := window.Element(event.Target()).Closest("#append-new-task"); !button.IsNull() {
			handleAppendTask(tmp, button)
			updateStateCounts()
		}
	})

	window.AddEventListenerFunc("change", func(event window.Event) {
		if checkbox := window.Element(event.Target()).Closest(taskFilterCheckboxSelector); !checkbox.IsNull() {
			handleToggleShowState(checkbox)
		}
	})

	select {}
}

func handleAppendTask(tmp *template.Template, button window.Element) {
	input := button.Closest(".new-task").QuerySelector("input#new-title")

	tasksEl := window.Document.QuerySelector(".tasks")

	task := todo.Task{
		ID:    primitive.NewObjectID(),
		Title: input.Get("value").String(),
		State: todo.TaskStateTODO,
	}

	taskEl, err := window.Document.NewElementFromTemplate(tmp, "task", task)
	if err != nil {
		fmt.Printf("failed to render task %s: %s", task.ID, err)
		return
	}

	input.Set("value", "")

	if tasksEl.ChildCount() == 0 {
		window.Document.QuerySelector(taskFilterCheckboxSelector+"[name=todo]").Set("checked", true)
	}

	showState := checkedStates()

	if !showState[task.State] {
		taskEl.Set("style", "display: none;")
	}

	tasksEl.AppendChild(taskEl)
}

func handleTaskTransition(tmp *template.Template, button window.Element) {
	taskEl := button.Closest(".task")

	task, err := taskFromElement(taskEl)
	if err != nil {
		fmt.Println(err)
		return
	}

	stateTransition := button.Attribute("data-state-transition")

	switch stateTransition {
	case "start":
		task.State, err = task.State.Start()
	case "delete":
		task.State, err = task.State.Delete()
	case "put-down":
		task.State, err = task.State.PutDown()
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

	if !checkedStates()[task.State] {
		newTaskEl.Set("style", "display: none;")
	}

	taskEl.Parent().ReplaceChild(newTaskEl, taskEl)
}

func handleToggleShowState(checkbox window.Element) {
	if checkbox.Attribute("id") == "all" && !checkbox.Get("checked").Truthy() {
		for _, c := range window.Document.QuerySelectorAll(taskFilterCheckboxSelector) {
			switch c.Attribute("name") {
			case "all":
			case "todo":
				c.Set("checked", true)
			default:
				c.Set("checked", false)
			}
		}
	}

	checkedStates := checkedStates()

	for _, taskEl := range window.Document.QuerySelectorAll(".task") {
		task, err := taskFromElement(taskEl)
		if err != nil {
			continue
		}

		if !checkedStates[task.State] {
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
	tasks := make([]todo.Task, 0)

	showState := checkedStates()

	tasksEL := window.Document.QuerySelector(".tasks")
	for _, task := range tasks {
		taskEl, err := window.Document.NewElementFromTemplate(tmp, "task", task)
		if err != nil {
			fmt.Printf("failed to render task %s: %s", task.ID, err)
			continue
		}

		if !showState[task.State] {
			taskEl.Set("style", "display: none;")
		}

		tasksEL.AppendChild(taskEl)
	}

	updateStateCounts()
}

func taskFromElement(el window.Element) (todo.Task, error) {
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

func checkedStates() map[todo.TaskState]bool {
	checkboxes := window.Document.QuerySelectorAll(taskFilterCheckboxSelector)

	states := make(map[todo.TaskState]bool, len(checkboxes))

	for _, checkbox := range checkboxes {
		states[todo.TaskState(checkbox.Attribute("name"))] = checkbox.Get("checked").Truthy()
	}

	if states["all"] {
		for s := range states {
			states[s] = true
		}
		for _, c := range checkboxes {
			c.Set("checked", true)
		}
	}

	return states
}

func updateStateCounts() {
	stateCounts := map[todo.TaskState]int{
		todo.TaskStateTODO:   0,
		todo.TaskStateActive: 0,
		todo.TaskStateReview: 0,
		todo.TaskStateDone:   0,
		"all":                0,
	}

	for _, taskEl := range window.Document.QuerySelectorAll(".task") {
		task, err := taskFromElement(taskEl)
		if err != nil {
			continue
		}

		stateCounts[task.State]++
		stateCounts["all"]++
	}

	checkboxSelector := `.task-filter>input[type="checkbox"][name=%q]`
	for state, count := range stateCounts {
		label := window.Document.QuerySelector(checkboxSelector+`+label`, state)

		if count == 0 {
			label.Set("style", "display: none;")
			continue
		}

		label.Set("style", "display: block;")

		if state == "all" {
			continue
		}

		countEl := label.QuerySelector(".count")
		countEl.SetInnerHTMLf("%d", count)
	}
}
