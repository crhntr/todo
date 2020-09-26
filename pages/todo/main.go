// +build js,wasm

package main

import (
	"fmt"
	"html/template"
	"syscall/js"

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

	updateStateCounts()

	window.AddEventListenerFunc("click", func(event window.Event) {
		if button := event.TargetElement().Closest(".task-transition"); button.Truthy() {
			handleTaskTransition(tmp, button)
			updateStateCounts()
		}

		if button := event.TargetElement().Closest("#append-new-task"); button.Truthy() {
			handleAppendTask(tmp, button)
			updateStateCounts()
		}
	})

	window.AddEventListenerFunc("change", func(event window.Event) {
		if checkbox := event.TargetElement().Closest(taskFilterCheckboxSelector); checkbox.Truthy() {
			handleToggleShowState(checkbox)
		}
	})

	window.AddEventListenerFunc("keyup", func(event window.Event) {
		if event.TargetElement().Matches("#new-title") {
			if event.KeyCode() != window.KeyCodeEnter {
				return
			}
			event.PreventDefault()
			window.Document.QuerySelector("#append-new-task").Call("click")
		}
	})

	handleTaskDragEvents()
}

func handleTaskDragEvents() {
	var (
		dragStart = make(chan window.Event)
		dragEnter = make(chan window.Event)
		dragEnd   = make(chan window.Event)

		taskBeingDragged, taskDraggedOver window.Element
	)

	window.AddEventListenerChannel("dragstart", dragStart)
	window.AddEventListenerChannel("dragenter", dragEnter)
	window.AddEventListenerChannel("dragend", dragEnd)

	for {
		select {
		case event := <-dragStart:

			if !event.Target().Truthy() {
				continue
			}

			el := window.Element(event.Target()).Closest(".task")
			if !el.Truthy() {
				continue
			}

			taskBeingDragged = el

		case event := <-dragEnter:

			if !event.Target().Truthy() {
				continue
			}

			el := window.Element(event.Target()).Closest(".task")
			if !el.Truthy() {
				continue
			}

			taskDraggedOver = el

		case event := <-dragEnd:

			if !event.Target().Truthy() ||
				!taskBeingDragged.Truthy() ||
				!taskDraggedOver.Truthy() {
				continue
			}

			d, o := taskBeingDragged, taskDraggedOver
			taskDraggedOver = window.Element(js.Null())
			taskDraggedOver = window.Element(js.Null())

			if d.Parent().IndexOf(d) > o.Parent().IndexOf(o) {
				o.InsertBefore(d)
				continue
			}

			o.InsertAfter(d)
		}
	}
}

func handleAppendTask(tmp *template.Template, button window.Element) {
	input := button.Closest(".new-task").QuerySelector("input#new-title")

	tasksEl := window.Document.QuerySelector(".tasks")

	task := todo.Task{
		ID:    primitive.NewObjectID(),
		Title: input.Get("value").String(),
		State: todo.TaskStateCreated,
	}

	taskEl, err := window.Document.NewElementFromTemplate(tmp, "task", task)
	if err != nil {
		fmt.Printf("failed to render task %s: %s", task.ID, err)
		return
	}

	input.Set("value", "")

	tasksEl.AppendChild(taskEl)
}

func handleTaskTransition(tmp *template.Template, button window.Element) {
	taskEl := button.Closest(".task")

	task, err := taskFromElement(taskEl)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch button.Attribute("data-state-transition") {
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
			case "created":
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
		State: todo.State(el.Attribute("data-state")),
		Title: el.QuerySelector(".title").Get("value").String(),
	}

	return task, nil
}

func checkedStates() map[todo.State]bool {
	checkboxes := window.Document.QuerySelectorAll(taskFilterCheckboxSelector)

	states := make(map[todo.State]bool, len(checkboxes))

	for _, checkbox := range checkboxes {
		states[todo.State(checkbox.Attribute("name"))] = checkbox.Get("checked").Truthy()
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
	stateCounts := map[todo.State]int{
		todo.TaskStateCreated: 0,
		todo.TaskStateActive:  0,
		todo.TaskStateReview:  0,
		todo.TaskStateDone:    0,
		"all":                 0,
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

		if countEl := label.QuerySelector(".count"); countEl.Truthy() {
			countEl.SetInnerHTMLf("%d", count)
		}
	}
}
