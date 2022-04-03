package trace

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io"
	"sync"
)

type ID string

func (i ID) String() string {
	return string(i)
}

type user struct {
	ids     map[string]bool
	isTrace bool
	m       sync.Mutex
	w       io.WriteCloser
}

type CreateWriteCloser interface {
	NewWriteCloser(fmt.Stringer) io.WriteCloser
}

type CreateReader interface {
	NewReader(fmt.Stringer) io.Reader
}

type users struct {
	users    sync.Map //map to user info
	creator  CreateWriteCloser
	useTrace bool
	err      error
}

func (u *user) String() string {
	var name string
	for i := range u.ids {
		name += fmt.Sprintf("%s", i)
	}
	return name
}

func (u *user) addUserID(ids ...string) *user {
	u.m.Lock()
	defer u.m.Unlock()
	for _, id := range ids {
		u.ids[id] = true
	}
	return u
}

func (u *users) newUser() *user {
	usr := &user{ids: make(map[string]bool)}
	return usr
}

func (u *users) getUser(id fmt.Stringer) *user {
	if usr := u.findUser(id.String()); usr != nil {
		return usr
	}
	return u.AddUser(id)
}

func (u *users) AddUser(ids ...fmt.Stringer) (usr *user) {
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = id.String()
	}
	defer func() {
		for _, id := range names {
			u.users.LoadOrStore(id, usr)
		}
		usr.addUserID(names...)
	}()
	if usr := u.findUser(names...); usr != nil {
		return usr
	}
	return u.newUser()
}

func (u *users) findUser(ids ...string) *user {
	for _, id := range ids {
		if usr, ok := u.users.Load(id); ok {
			return usr.(*user)
		}
	}
	return nil
}

func (u *users) Write(data []io.Reader, id fmt.Stringer) error {
	usr := u.getUser(id)
	if u.useTrace && !usr.isTrace {
		return nil
	}
	usr.m.Lock()
	defer usr.m.Unlock()
	if usr.w == nil {
		usr.w = u.creator.NewWriteCloser(usr)
	}
	for _, r := range data {
		if _, err := io.Copy(usr.w, r); err != nil {
			return err
		}
	}
	return nil
}

func (u *users) Close() error {
	var e error
	u.users.Range(func(key, value interface{}) bool {
		if value.(*user).w == nil {
			return true
		}
		if err := value.(*user).w.Close(); err != nil {
			e = err
		}
		value.(*user).w = nil
		return true
	})
	return e
}

func (u *users) Trace(id fmt.Stringer, isTrace bool) *users {
	u.findUser(id.String()).isTrace = isTrace
	return u
}

func (u *users) UseTrace(isUse bool) *users {
	u.useTrace = isUse
	return u
}

func (u *users) DelUser(id fmt.Stringer) {
	if usr := u.findUser(id.String()); usr != nil {
		usr.w.Close()
		for i := range usr.ids {
			u.users.Delete(i)
		}
	}
}

func (u *users) GetUsrInfo(id fmt.Stringer) string {
	usr := u.findUser(id.String())
	return spew.Sdump(usr.w, "isTrace:", usr.isTrace)
}

func (u *users) ListUsers() string {
	var users []string
	u.users.Range(func(key, value interface{}) bool {
		users = append(users, key.(string))
		return true
	})
	return spew.Sdump("isUseTrace:", u.useTrace, users)
}
