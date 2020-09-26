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
	TaskStateTODO      TaskState = "todo"
	TaskStateActive    TaskState = "active"
	TaskStateCompleted TaskState = "completed"
	TaskStateDone      TaskState = "done"
	TaskStateRedo      TaskState = "redo"
)

const (
	errTaskStateTransitionFailedFormat = "task transition %s failed because the current task state is %s"
)

func (state TaskState) Start() (TaskState, error) {
	switch state {
	case TaskStateTODO, TaskStateRedo:
		return TaskStateActive, nil
	default:
		return "", fmt.Errorf(errTaskStateTransitionFailedFormat, "start", state)
	}
}

func (state TaskState) Finish() (TaskState, error) {
	switch state {
	case TaskStateActive:
		return TaskStateCompleted, nil
	default:
		return "", fmt.Errorf(errTaskStateTransitionFailedFormat, "finish", state)
	}
}

func (state TaskState) Review(passed bool) (TaskState, error) {
	switch state {
	case TaskStateCompleted:
		if passed {
			return TaskStateDone, nil
		}
		return TaskStateRedo, nil
	default:
		return "", fmt.Errorf(errTaskStateTransitionFailedFormat, "review", state)
	}
}
