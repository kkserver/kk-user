package user

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type UserLoginTaskResult struct {
	app.Result
	User *User `json:"user,omitempty"`
}

type UserLoginTask struct {
	app.Task
	Name     string `json:"name"`
	Password string `json:"password"`
	Result   UserLoginTaskResult
}

func (task *UserLoginTask) GetResult() interface{} {
	return &task.Result
}

func (task *UserLoginTask) GetInhertType() string {
	return "user"
}

func (task *UserLoginTask) GetClientName() string {
	return "User.Login"
}
