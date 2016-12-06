package user

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type UserSetTaskResult struct {
	app.Result
	User *User `json:"user,omitempty"`
}

type UserSetTask struct {
	app.Task
	Uid      int64  `json:"uid"`
	Password string `json:"password"`
	Result   UserSetTaskResult
}

func (task *UserSetTask) GetResult() interface{} {
	return &task.Result
}

func (task *UserSetTask) GetInhertType() string {
	return "user"
}

func (task *UserSetTask) GetClientName() string {
	return "User.Set"
}
