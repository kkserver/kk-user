package user

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type UserOptionsTaskResult struct {
	app.Result
	Options interface{} `json:"options,omitempty"`
}

type UserOptionsTask struct {
	app.Task
	Uid    int64  `json:"uid"`
	Name   string `json:"name"`
	Result UserOptionsTaskResult
}

func (task *UserOptionsTask) GetResult() interface{} {
	return &task.Result
}

func (task *UserOptionsTask) GetInhertType() string {
	return "user"
}

func (task *UserOptionsTask) GetClientName() string {
	return "User.GetOptions"
}
