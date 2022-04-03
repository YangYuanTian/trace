package trace

import (
	"fmt"
	"io"
	"sync"
)

type user struct {
	key uint64
	ids map[string]bool
	m   sync.Mutex
	w   io.WriteCloser
}

type CreateWriteCloser interface {
	NewWriteCloser(fmt.Stringer) io.WriteCloser
}
type CreateReader interface {
	NewReader(fmt.Stringer) io.Reader
}

type users struct {
	users    sync.Map //map to user info
	ids      sync.Map //map to key
	key      uint64
	m        sync.Mutex
	start    bool
	creator  CreateWriteCloser
	err      error
}

func (u *user) String() string {
	var name string
	for i := range u.ids {
		name += fmt.Sprintf("%s", i)
	}
	return name
}

func (u *users) newKey(ids ...string) uint64 {
	u.m.Lock()
	defer u.m.Unlock()
	u.key++
	for i := range ids {
		u.ids.Store(ids[i], u.key)
	}
	return u.key
}

func (u *users) findKey(ids ...string) (uint64, bool) {
	for  _,id := range ids {
		if key, ok := u.ids.Load(id); ok {
			return key.(uint64), true
		}
	}
	return 0, false
}

func (u *users) findUser(key uint64) *user {
	if usr, ok := u.users.Load(key); ok {
		return usr.(*user)
	}
	usr := &user{
		key: key,
		ids: make(map[string]bool),
	}
	u.users.Store(key, usr)
	usr.w = u.creator.NewWriteCloser(usr)
	return usr
}

func (u *users) SetNames(ids ...fmt.Stringer) {
	if len(ids) == 0 {
		return
	}
	names := make([]string, len(ids))
	for i := range ids {
		names[i] = ids[i].String()
	}
	key, ok := u.findKey(names...)
	switch !ok {
	case true:
		key = u.newKey(names...)
	case false:
		for i := range names {
			u.ids.Store(names[i], key)
		}
	}
	u.findUser(key).addName(names...)
}

func (u *user) addName(ids ...string) *user {
	u.m.Lock()
	defer u.m.Unlock()
	for i := range ids {
		u.ids[ids[i]] = true
	}
	return u
}

func (u *users) getInfo(id string) *user {
	if key, ok := u.findKey(id); ok {
		return u.findUser(key)
	}
	return u.findUser(u.newKey(id)).addName(id)
}

func (u *users) Write(data []io.Reader, id fmt.Stringer) error {
	usr := u.getInfo(id.String())
	usr.m.Lock()
	defer usr.m.Unlock()
	for _, r := range data {
		if _, err := io.Copy(usr.w, r); err != nil {
			return err
		}
	}
	return nil
}

func (u *users) Close() error {
	var err error
	u.users.Range(func(key, value interface{}) bool {
		err = value.(*user).w.Close()
		return true
	})
	return err
}
