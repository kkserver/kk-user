package user

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type UserQueryCounter struct {
	PageIndex int `json:"p"`
	PageSize  int `json:"size"`
	PageCount int `json:"count"`
	RowCount  int `json:"rowCount"`
}

type UserQueryTaskResult struct {
	app.Result
	Counter *UserQueryCounter `json:"counter,omitempty"`
	Users   []User            `json:"users,omitempty"`
}

type UserQueryTask struct {
	app.Task
	Uid       int64  `json:"uid"`
	Name      string `json:"name"`
	OrderBy   string `json:"orderBy"` // desc, asc
	PageIndex int    `json:"p"`
	PageSize  int    `json:"size"`
	Counter   bool   `json:"counter"`
	Result    UserQueryTaskResult
}

func (T *UserQueryTask) GetResult() interface{} {
	return &T.Result
}

func (T *UserQueryTask) GetInhertType() string {
	return "user"
}

func (T *UserQueryTask) GetClientName() string {
	return "User.Query"
}
