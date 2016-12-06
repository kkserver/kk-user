package user

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type UserSetOptionsTaskResult struct {
	app.Result
}

type UserSetOptionsTask struct {
	app.Task
	Uid     int64       `json:"uid"`
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Options interface{} `json:"options"`
	Result  UserSetOptionsTaskResult
}

func (task *UserSetOptionsTask) GetResult() interface{} {
	return &task.Result
}

func (task *UserSetOptionsTask) GetInhertType() string {
	return "user"
}

func (task *UserSetOptionsTask) GetClientName() string {
	return "User.SetOptions"
}
