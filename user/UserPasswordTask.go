package user

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type UserPasswordTaskResult struct {
	app.Result
	User *User `json:"user,omitempty"`
}

type UserPasswordTask struct {
	app.Task
	Uid      int64  `json:"uid"`
	Password string `json:"password"`
	Result   UserPasswordTaskResult
}

func (task *UserPasswordTask) GetResult() interface{} {
	return &task.Result
}

func (task *UserPasswordTask) GetInhertType() string {
	return "user"
}

func (task *UserPasswordTask) GetClientName() string {
	return "User.Password"
}
