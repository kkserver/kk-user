package user

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/app"
	"github.com/kkserver/kk-lib/kk/app/client"
	"github.com/kkserver/kk-lib/kk/app/remote"
	"github.com/kkserver/kk-lib/kk/json"
	Value "github.com/kkserver/kk-lib/kk/value"
	"math/rand"
	"reflect"
	"time"
)

const UserOptionsTypeText = "text"
const UserOptionsTypeJson = "json"

type User struct {
	Id       int64  `json:"id"`
	Phone    string `json:"phone"`
	Password string `json:"-"`
	Ctime    int64  `json:"ctime"`
	Atime    int64  `json:"atime"`
	Mtime    int64  `json:"mtime"`
}

type UserOptions struct {
	Id      int64  `json:"id"`
	Uid     int64  `json:"uid"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Options string `json:"options"`
}

type UserApp struct {
	app.App
	DB          *app.DBConfig
	User        *UserService
	Remote      *remote.Service
	Client      *client.Service
	ClientCache *client.WithService

	Token   string
	Expires int64

	UserTable        kk.DBTable
	UserOptionsTable kk.DBTable
}

func (C *UserApp) GetDB() (*sql.DB, error) {
	return C.DB.Get(C)
}

func EncodePassword(a *UserApp, password string) string {
	m := md5.New()
	m.Write([]byte(password))
	m.Write([]byte(a.Token))
	v := m.Sum(nil)
	return hex.EncodeToString(v)
}

func NewPassword(a *UserApp) string {
	return EncodePassword(a, fmt.Sprintf("%d %d", time.Now().UnixNano(), rand.Intn(100000)))
}

func (U *UserOptions) GetOptions() interface{} {

	if U.Type == UserOptionsTypeJson {

		if U.Options == "" {
			return nil
		}

		var object interface{} = nil

		var err = json.Decode([]byte(U.Options), &object)

		if err == nil {
			return object
		}

		return nil
	}

	return U.Options
}

func (U *UserOptions) SetOptions(options interface{}) {

	if U.Type == UserOptionsTypeJson {

		var object = U.GetOptions()

		if object == nil {
			object = map[string]interface{}{}
		}

		if options != nil {

			var m, ok = options.(map[string]interface{})
			var v = reflect.ValueOf(object)

			if ok {

				for key, value := range m {
					Value.Set(v, key, reflect.ValueOf(value))
				}

			}
		}

		b, _ := json.Encode(object)
		U.Options = string(b)
	} else {
		U.Options = Value.StringValue(reflect.ValueOf(options), "")
	}

}
