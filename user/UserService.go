package user

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/kkserver/kk-cache/cache"
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/app"
	"github.com/kkserver/kk-lib/kk/dynamic"
	"github.com/kkserver/kk-lib/kk/json"
	"log"
	"strings"
	"time"
)

type UserService struct {
	app.Service

	Init       *app.InitTask
	Create     *UserCreateTask
	Get        *UserTask
	Set        *UserSetTask
	Login      *UserLoginTask
	Password   *UserPasswordTask
	GetOptions *UserOptionsTask
	SetOptions *UserSetOptionsTask
	Query      *UserQueryTask

	Users map[string]interface{} //初始化用户
}

func (S *UserService) Handle(a app.IApp, task app.ITask) error {
	return app.ServiceReflectHandle(a, task, S)
}

func (S *UserService) HandleInitTask(a *UserApp, task *app.InitTask) error {

	db, err := a.GetDB()

	if err != nil {
		log.Println("[UserService][HandleInitTask]" + err.Error())
		return nil
	}

	v := User{}

	if S.Users != nil {

		for name, password := range S.Users {

			rows, err := kk.DBQuery(db, &a.UserTable, a.DB.Prefix, " WHERE name=?", name)

			if err == nil {

				if !rows.Next() {

					v.Name = name
					v.Password = EncodePassword(a, dynamic.StringValue(password, ""))
					v.Atime = time.Now().Unix()
					v.Mtime = v.Atime
					v.Ctime = v.Atime

					_, err = kk.DBInsert(db, &a.UserTable, a.DB.Prefix, &v)

					if err != nil {
						log.Println(err)
					} else {
						log.Println("Create User " + name)
					}
				}

				rows.Close()

			} else {
				log.Println(err)
			}
		}

	}

	return nil
}

func (S *UserService) HandleUserCreateTask(a *UserApp, task *UserCreateTask) error {

	if task.Name == "" {
		task.Result.Errno = ERROR_USER_NOT_FOUND_NAME
		task.Result.Errmsg = "Not found name"
		return nil
	}

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var prefix = a.DB.Prefix

	tx, err := db.Begin()

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	func() {

		var v = User{}

		rows, err := kk.DBQuery(db, &a.UserTable, prefix, " WHERE name=?", task.Name)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return
		}

		defer rows.Close()

		if rows.Next() {
			task.Result.Errno = ERROR_USER_NAME
			task.Result.Errmsg = "The name already exists"
			return
		}

		v.Name = task.Name

		if task.Password == "" {
			v.Password = NewPassword(a)
		} else {
			v.Password = EncodePassword(a, task.Password)
		}

		v.Atime = time.Now().Unix()
		v.Mtime = v.Atime
		v.Ctime = v.Atime

		_, err = kk.DBInsert(db, &a.UserTable, prefix, &v)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return
		}

		task.Result.User = &v

	}()

	if task.Result.Errno != 0 {
		tx.Rollback()
		return nil
	}

	err = tx.Commit()

	if err != nil {
		tx.Rollback()
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	return nil
}

func (S *UserService) HandleUserSetTask(a *UserApp, task *UserSetTask) error {

	if task.Uid == 0 {
		task.Result.Errno = ERROR_USER_NOT_FOUND_UID
		task.Result.Errmsg = "Not found uid"
		return nil
	}

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var prefix = a.DB.Prefix
	var v = User{}
	var scanner = kk.NewDBScaner(&v)

	rows, err := kk.DBQuery(db, &a.UserTable, prefix, " WHERE id=?", task.Uid)

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	defer rows.Close()

	if rows.Next() {

		err = scanner.Scan(rows)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}

		if task.Password == "" {
			v.Password = NewPassword(a)
		} else {
			v.Password = EncodePassword(a, task.Password)
		}

		v.Mtime = time.Now().Unix()

		_, err = kk.DBUpdateWithKeys(db, &a.UserTable, prefix, &v, map[string]bool{"password": true, "mtime": true})

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}

		task.Result.User = &v

	} else {
		task.Result.Errno = ERROR_USER_NOT_FOUND
		task.Result.Errmsg = "Not found user"
	}

	return nil
}

func (S *UserService) HandleUserTask(a *UserApp, task *UserTask) error {

	if task.Uid == 0 && task.Name == "" {
		task.Result.Errno = ERROR_USER_NOT_FOUND_UID
		task.Result.Errmsg = "Not found uid"
		return nil
	}

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var prefix = a.DB.Prefix
	var v = User{}
	var scanner = kk.NewDBScaner(&v)
	var rows *sql.Rows = nil

	if task.Uid != 0 {
		rows, err = kk.DBQuery(db, &a.UserTable, prefix, " WHERE id=?", task.Uid)
	} else {
		rows, err = kk.DBQuery(db, &a.UserTable, prefix, " WHERE name=?", task.Name)
	}

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	defer rows.Close()

	if rows.Next() {

		err = scanner.Scan(rows)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}

		task.Result.User = &v

	} else {

		if task.Autocreate && task.Name != "" {
			var create = UserCreateTask{}
			create.Name = task.Name
			app.Handle(a, &create)
			if create.Result.Errno == 0 && create.Result.User != nil {
				task.Result.User = create.Result.User
				return nil
			}
		}

		task.Result.Errno = ERROR_USER_NOT_FOUND
		task.Result.Errmsg = "Not found user"
	}

	return nil
}

func (S *UserService) HandleUserOptionsTask(a *UserApp, task *UserOptionsTask) error {

	if task.Uid == 0 {
		task.Result.Errno = ERROR_USER_NOT_FOUND_UID
		task.Result.Errmsg = "Not found uid"
		return nil
	}

	var key = fmt.Sprintf("%s.%d.%s", a.CacheKey, task.Uid, task.Name)

	{
		var cache = cache.CacheTask{}
		cache.Key = key
		var err = app.Handle(a, &cache)
		if err == nil && cache.Result.Errno == 0 && cache.Result.Value != "" {
			var vv = UserOptions{}
			err = json.Decode([]byte(cache.Result.Value), &vv)
			if err == nil {
				task.Result.Options = vv.GetOptions()
				return nil
			}
		}
	}

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var prefix = a.DB.Prefix

	var v = UserOptions{}
	var scanner = kk.NewDBScaner(&v)

	rows, err := kk.DBQuery(db, &a.UserOptionsTable, prefix, " WHERE uid=? AND name=?", task.Uid, task.Name)

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	defer rows.Close()

	if rows.Next() {

		err = scanner.Scan(rows)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}

		task.Result.Options = v.GetOptions()

		{
			var cache = cache.CacheSetTask{}
			cache.Key = key
			cache.Expires = a.Expires
			b, _ := json.Encode(&v)
			cache.Value = string(b)
			app.Handle(a, &cache)
		}
	}

	return nil
}

func (S *UserService) HandleUserSetOptionsTask(a *UserApp, task *UserSetOptionsTask) error {

	if task.Uid == 0 {
		task.Result.Errno = ERROR_USER_NOT_FOUND_UID
		task.Result.Errmsg = "Not found uid"
		return nil
	}

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var prefix = a.DB.Prefix

	var v = UserOptions{}
	var scanner = kk.NewDBScaner(&v)

	rows, err := kk.DBQuery(db, &a.UserOptionsTable, prefix, " WHERE uid=? AND name=?", task.Uid, task.Name)

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	defer rows.Close()

	if rows.Next() {

		err = scanner.Scan(rows)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}

		if task.Type != v.Type {
			v.Type = task.Type
			v.Options = ""
		}

		v.SetOptions(task.Options)

		_, err = db.Exec(fmt.Sprintf("UPDATE %s%s SET mtime=? WHERE id=?", prefix, a.UserTable.Name), time.Now().Unix(), task.Uid)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}

		_, err = kk.DBUpdateWithKeys(db, &a.UserOptionsTable, prefix, &v, map[string]bool{"options": true, "type": true})

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}

	} else {

		v.Type = task.Type
		v.Uid = task.Uid
		v.Name = task.Name
		v.SetOptions(task.Options)

		_, err = db.Exec(fmt.Sprintf("UPDATE %s%s SET mtime=? WHERE id=?", prefix, a.UserTable.Name), time.Now().Unix(), task.Uid)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}

		_, err = kk.DBInsert(db, &a.UserOptionsTable, prefix, &v)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}
	}

	{
		var cache = cache.CacheRemoveTask{}
		cache.Key = fmt.Sprintf("%s.%d.%s", a.CacheKey, v.Uid, v.Name)
		app.Handle(a, &cache)
	}

	return nil
}

func (S *UserService) HandleUserLoginTask(a *UserApp, task *UserLoginTask) error {

	if task.Name == "" {
		task.Result.Errno = ERROR_USER_NOT_FOUND_NAME
		task.Result.Errmsg = "Not found name"
		return nil
	}

	if task.Password == "" {
		task.Result.Errno = ERROR_USER_NOT_FOUND_PASSWORD
		task.Result.Errmsg = "Not found password"
		return nil
	}

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var prefix = a.DB.Prefix
	var v = User{}
	var scanner = kk.NewDBScaner(&v)

	rows, err := kk.DBQuery(db, &a.UserTable, prefix, " WHERE name=?", task.Name)

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	defer rows.Close()

	if rows.Next() {

		err = scanner.Scan(rows)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}

		if EncodePassword(a, task.Password) != v.Password {
			task.Result.Errno = ERROR_USER_PASSWORD
			task.Result.Errmsg = "user password fail"
			return nil
		}

		v.Atime = time.Now().Unix()

		_, err = kk.DBUpdateWithKeys(db, &a.UserTable, prefix, &v, map[string]bool{"atime": true})

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}

		task.Result.User = &v

	} else {
		task.Result.Errno = ERROR_USER_NOT_FOUND
		task.Result.Errmsg = "Not found user"
	}

	return nil
}

func (S *UserService) HandleUserPasswordTask(a *UserApp, task *UserPasswordTask) error {

	if task.Uid == 0 {
		task.Result.Errno = ERROR_USER_NOT_FOUND_UID
		task.Result.Errmsg = "Not found uid"
		return nil
	}

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var prefix = a.DB.Prefix
	var v = User{}
	var scanner = kk.NewDBScaner(&v)

	rows, err := kk.DBQuery(db, &a.UserTable, prefix, " WHERE id=?", task.Uid)

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	defer rows.Close()

	if rows.Next() {

		err = scanner.Scan(rows)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}

		if EncodePassword(a, task.Password) != v.Password {
			task.Result.Errno = ERROR_USER_PASSWORD
			task.Result.Errmsg = "user password fail"
			return nil
		}

		task.Result.User = &v

	} else {
		task.Result.Errno = ERROR_USER_NOT_FOUND
		task.Result.Errmsg = "Not found user"
	}

	return nil
}

func (S *UserService) HandleUserQueryTask(a *UserApp, task *UserQueryTask) error {

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var users = []User{}
	var prefix = a.DB.Prefix

	sql := bytes.NewBuffer(nil)

	args := []interface{}{}

	sql.WriteString(" WHERE 1")

	if task.Uid != 0 {
		sql.WriteString(" AND id=?")
		args = append(args, task.Uid)
	}

	if task.Name != "" {
		sql.WriteString(" AND name=?")
		args = append(args, task.Name)
	}

	if task.Names != "" {

		sql.WriteString(" AND name IN (")
		for i, v := range strings.Split(task.Names, ",") {
			if i != 0 {
				sql.WriteString(",")
			}
			sql.WriteString("?")
			args = append(args, v)
		}
		sql.WriteString(")")

	}

	if task.OrderBy == "asc" {
		sql.WriteString(" ORDER BY id ASC")
	} else {
		sql.WriteString(" ORDER BY id DESC")
	}

	var pageIndex = task.PageIndex
	var pageSize = task.PageSize

	if pageIndex < 1 {
		pageIndex = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	if task.Counter {
		var counter = UserQueryCounter{}
		counter.PageIndex = pageIndex
		counter.PageSize = pageSize
		counter.RowCount, err = kk.DBQueryCount(db, &a.UserTable, prefix, sql.String(), args...)
		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}
		if counter.RowCount%pageSize == 0 {
			counter.PageCount = counter.RowCount / pageSize
		} else {
			counter.PageCount = counter.RowCount/pageSize + 1
		}
		task.Result.Counter = &counter
	}

	sql.WriteString(fmt.Sprintf(" LIMIT %d,%d", (pageIndex-1)*pageSize, pageSize))

	var v = User{}
	var scanner = kk.NewDBScaner(&v)

	rows, err := kk.DBQuery(db, &a.UserTable, prefix, sql.String(), args...)

	if err != nil {
		task.Result.Errno = ERROR_USER
		task.Result.Errmsg = err.Error()
		return nil
	}

	defer rows.Close()

	for rows.Next() {

		err = scanner.Scan(rows)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return nil
		}

		users = append(users, v)
	}

	task.Result.Users = users

	return nil
}
