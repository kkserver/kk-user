package user

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type UserTaskResult struct {
	app.Result
	User *User `json:"user,omitempty"`
}

type UserTask struct {
	app.Task
	Uid        int64  `json:"uid"`
	Name       string `json:"name"`
	Autocreate bool   `json:"autocreate"`
	Result     UserTaskResult
}

func (task *UserTask) GetResult() interface{} {
	return &task.Result
}

func (task *UserTask) GetInhertType() string {
	return "user"
}

func (task *UserTask) GetClientName() string {
	return "User.Get"
}
