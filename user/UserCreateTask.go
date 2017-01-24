package user

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type UserCreateTaskResult struct {
	app.Result
	User *User `json:"user,omitempty"`
}

type UserCreateTask struct {
	app.Task
	Name     string `json:"name"`
	Password string `json:"password"`
	Result   UserCreateTaskResult
}

func (task *UserCreateTask) GetResult() interface{} {
	return &task.Result
}

func (task *UserCreateTask) GetInhertType() string {
	return "user"
}

func (task *UserCreateTask) GetClientName() string {
	return "User.Create"
}
