package user

import (
	"database/sql"
	"fmt"
	"github.com/kkserver/kk-cache/cache"
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/app"
	"github.com/kkserver/kk-lib/kk/json"
	"log"
	"time"
)

type UserService struct {
	app.Service
	Init       *app.InitTask
	Create     *UserCreateTask
	Get        *UserTask
	Set        *UserSetTask
	Login      *UserLoginTask
	GetOptions *UserOptionsTask
	SetOptions *UserSetOptionsTask
}

func (S *UserService) Handle(a app.IApp, task app.ITask) error {
	return app.ServiceReflectHandle(a, task, S)
}

func (S *UserService) HandleInitTask(a *UserApp, task *app.InitTask) error {
	_, err := a.GetDB()
	if err != nil {
		log.Println("[UserService][HandleInitTask]" + err.Error())
	}
	return nil
}

func (S *UserService) HandleUserCreateTask(a *UserApp, task *UserCreateTask) error {

	if task.Phone == "" {
		task.Result.Errno = ERROR_USER_NOT_FOUND_PHONE
		task.Result.Errmsg = "Not found phone"
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

		rows, err := kk.DBQuery(db, &a.UserTable, prefix, " WHERE phone=?", task.Phone)

		if err != nil {
			task.Result.Errno = ERROR_USER
			task.Result.Errmsg = err.Error()
			return
		}

		defer rows.Close()

		if rows.Next() {
			task.Result.Errno = ERROR_USER_PHONE
			task.Result.Errmsg = "Phone number already exists"
			return
		}

		v.Phone = task.Phone

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

	if task.Uid == 0 && task.Phone == "" {
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
		rows, err = kk.DBQuery(db, &a.UserTable, prefix, " WHERE phone=?", task.Phone)
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

		if task.Autocreate && task.Phone != "" {
			var create = UserCreateTask{}
			create.Phone = task.Phone
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

	var key = fmt.Sprintf("user.options.%d.%s", task.Uid, task.Name)

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

	return nil
}

func (S *UserService) HandleUserLoginTask(a *UserApp, task *UserLoginTask) error {

	if task.Phone == "" {
		task.Result.Errno = ERROR_USER_NOT_FOUND_PHONE
		task.Result.Errmsg = "Not found phone"
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

	rows, err := kk.DBQuery(db, &a.UserTable, prefix, " WHERE phone=?", task.Phone)

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
