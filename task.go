package todo

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	ID    primitive.ObjectID `json:"id"`
	Title string             `json:"task"`
	State `json:"state"`
}

type State string

const (
	TaskStateTODO    State = "todo"
	TaskStateActive  State = "active"
	TaskStateReview  State = "finished"
	TaskStateDone    State = "done"
	TaskStateDeleted State = "deleted"
)

const (
	errTaskStateTransitionFailedFormat = "task transition %s failed because the current task state is %s"
)

func (state State) CanStart() bool {
	switch state {
	case TaskStateTODO:
		return true
	default:
		return false
	}
}

func (state State) Start() (State, error) {
	if !state.CanStart() {
		return "", fmt.Errorf(errTaskStateTransitionFailedFormat, "start", state)
	}

	return TaskStateActive, nil
}

func (state State) CanPutDown() bool {
	return state == TaskStateActive
}

func (state State) PutDown() (State, error) {
	if !state.CanPutDown() {
		return "", fmt.Errorf(errTaskStateTransitionFailedFormat, "put down", state)
	}

	return TaskStateTODO, nil
}

func (state State) CanFinish() bool {
	switch state {
	case TaskStateActive:
		return true
	default:
		return false
	}
}

func (state State) Finish() (State, error) {
	if !state.CanFinish() {
		return "", fmt.Errorf(errTaskStateTransitionFailedFormat, "finish", state)
	}

	return TaskStateReview, nil
}

func (state State) CanReview() bool {
	switch state {
	case TaskStateReview:
		return true
	default:

		return false
	}
}

func (state State) IsDone() bool {
	return state == TaskStateDone
}

func (state State) Review(passed bool) (State, error) {
	if !state.CanReview() {
		return "", fmt.Errorf(errTaskStateTransitionFailedFormat, "review", state)
	}

	if passed {
		return TaskStateDone, nil
	}

	return TaskStateTODO, nil
}

func (state State) CanDelete() bool {
	return state == TaskStateTODO
}

func (state State) Delete() (State, error) {
	if !state.CanDelete() {
		return "", fmt.Errorf(errTaskStateTransitionFailedFormat, "delete", state)
	}

	return TaskStateDeleted, nil
}
