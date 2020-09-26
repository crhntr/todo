package todo

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	ID    primitive.ObjectID `json:"id"`
	Title string             `json:"task"`
	State TaskState          `json:"state"`
}

type TaskState string

const (
	TaskStateTODO    TaskState = "todo"
	TaskStateActive  TaskState = "active"
	TaskStateReview  TaskState = "finished"
	TaskStateDone    TaskState = "done"
	TaskStateDeleted TaskState = "deleted"
)

const (
	errTaskStateTransitionFailedFormat = "task transition %s failed because the current task state is %s"
)

func (state TaskState) CanStart() bool {
	switch state {
	case TaskStateTODO:
		return true
	default:
		return false
	}
}

func (state TaskState) Start() (TaskState, error) {
	if !state.CanStart() {
		return "", fmt.Errorf(errTaskStateTransitionFailedFormat, "start", state)
	}

	return TaskStateActive, nil
}

func (state TaskState) CanFinish() bool {
	switch state {
	case TaskStateActive:
		return true
	default:
		return false
	}
}

func (state TaskState) Finish() (TaskState, error) {
	if !state.CanFinish() {
		return "", fmt.Errorf(errTaskStateTransitionFailedFormat, "finish", state)
	}

	return TaskStateReview, nil
}

func (state TaskState) CanReview() bool {
	switch state {
	case TaskStateReview:
		return true
	default:

		return false
	}
}

func (state TaskState) IsDone() bool {
	return state == TaskStateDone
}

func (state TaskState) Review(passed bool) (TaskState, error) {
	if !state.CanReview() {
		return "", fmt.Errorf(errTaskStateTransitionFailedFormat, "review", state)
	}

	if passed {
		return TaskStateDone, nil
	}

	return TaskStateTODO, nil
}

func (state TaskState) CanDelete() bool {
	return state == TaskStateTODO
}

func (state TaskState) Delete() (TaskState, error) {
	if !state.CanDelete() {
		return "", fmt.Errorf(errTaskStateTransitionFailedFormat, "delete", state)
	}

	return TaskStateDeleted, nil
}
