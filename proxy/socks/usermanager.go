package socks

import (
	"context"
	"crypto/subtle"
	"strings"
	"sync"

	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/proxy"
)

// userValidator stores SOCKS users in memory. Accounts loaded from the static
// config are seeded into the same registry so that they remain available to
// the dynamic authentication path.
type userValidator struct {
	mu      sync.RWMutex
	byName  map[string]*protocol.MemoryUser
	byEmail map[string]*protocol.MemoryUser
	static  map[string]struct{}
}

func newUserValidator() *userValidator {
	return &userValidator{
		byName:  make(map[string]*protocol.MemoryUser),
		byEmail: make(map[string]*protocol.MemoryUser),
		static:  make(map[string]struct{}),
	}
}

func (v *userValidator) seed(accounts map[string]string, level uint32) error {
	for username, password := range accounts {
		user := &protocol.MemoryUser{
			Email: username,
			Level: level,
			Account: &Account{
				Username: username,
				Password: password,
			},
		}
		if err := v.add(user, false, true); err != nil {
			return err
		}
	}
	return nil
}

func (v *userValidator) add(u *protocol.MemoryUser, replaceStatic, markStatic bool) error {
	if u == nil {
		return errors.New("socks: user must not be nil")
	}
	if u.Email == "" {
		return errors.New("socks: user email must not be empty")
	}
	account, ok := u.Account.(*Account)
	if !ok || account == nil {
		return errors.New("socks: account must be *socks.Account")
	}
	if account.Username == "" {
		return errors.New("socks: username must not be empty")
	}
	if account.Password == "" {
		return errors.New("socks: password must not be empty")
	}

	emailKey := strings.ToLower(u.Email)

	v.mu.Lock()
	defer v.mu.Unlock()

	existingByName, foundByName := v.byName[account.Username]
	existingByEmail, foundByEmail := v.byEmail[emailKey]
	if foundByName && (!replaceStatic || !v.isStaticLocked(existingByName)) {
		return errors.New("socks: user ", account.Username, " already exists")
	}
	if foundByEmail && (!replaceStatic || !v.isStaticLocked(existingByEmail)) {
		return errors.New("socks: email ", u.Email, " already exists")
	}
	if foundByName {
		v.deleteLocked(existingByName)
	}
	if foundByEmail {
		v.deleteLocked(existingByEmail)
	}

	v.byName[account.Username] = u
	v.byEmail[emailKey] = u
	if markStatic {
		v.static[account.Username] = struct{}{}
	}
	return nil
}

func (v *userValidator) isStaticLocked(u *protocol.MemoryUser) bool {
	account, ok := u.Account.(*Account)
	if !ok || account == nil {
		return false
	}
	_, ok = v.static[account.Username]
	return ok
}

func (v *userValidator) deleteLocked(u *protocol.MemoryUser) {
	if u == nil {
		return
	}
	if account, ok := u.Account.(*Account); ok && account != nil {
		delete(v.byName, account.Username)
		delete(v.static, account.Username)
	}
	delete(v.byEmail, strings.ToLower(u.Email))
}

// Add adds a dynamic user. It replaces a user that was seeded from the static
// config with the same username or email.
func (v *userValidator) Add(u *protocol.MemoryUser) error {
	return v.add(u, true, false)
}

func (v *userValidator) Del(email string) error {
	if email == "" {
		return errors.New("socks: email must not be empty")
	}

	emailKey := strings.ToLower(email)

	v.mu.Lock()
	defer v.mu.Unlock()

	u, found := v.byEmail[emailKey]
	if !found {
		return errors.New("socks: user ", email, " not found")
	}
	v.deleteLocked(u)
	return nil
}

func (v *userValidator) Authenticate(username, password string) (*protocol.MemoryUser, bool) {
	v.mu.RLock()
	u, found := v.byName[username]
	v.mu.RUnlock()
	if !found {
		return nil, false
	}

	account, ok := u.Account.(*Account)
	if !ok || account == nil {
		return nil, false
	}
	if subtle.ConstantTimeCompare([]byte(account.Password), []byte(password)) != 1 {
		return nil, false
	}
	return u, true
}

func (v *userValidator) GetByEmail(email string) *protocol.MemoryUser {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.byEmail[strings.ToLower(email)]
}

func (v *userValidator) GetAll() []*protocol.MemoryUser {
	v.mu.RLock()
	defer v.mu.RUnlock()
	users := make([]*protocol.MemoryUser, 0, len(v.byEmail))
	for _, u := range v.byEmail {
		users = append(users, u)
	}
	return users
}

func (v *userValidator) GetCount() int64 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return int64(len(v.byEmail))
}

// AddUser implements proxy.UserManager.AddUser().
func (s *Server) AddUser(ctx context.Context, u *protocol.MemoryUser) error {
	return s.validator.Add(u)
}

// RemoveUser implements proxy.UserManager.RemoveUser().
func (s *Server) RemoveUser(ctx context.Context, email string) error {
	return s.validator.Del(email)
}

// GetUser implements proxy.UserManager.GetUser().
func (s *Server) GetUser(ctx context.Context, email string) *protocol.MemoryUser {
	return s.validator.GetByEmail(email)
}

// GetUsers implements proxy.UserManager.GetUsers().
func (s *Server) GetUsers(ctx context.Context) []*protocol.MemoryUser {
	return s.validator.GetAll()
}

// GetUsersCount implements proxy.UserManager.GetUsersCount().
func (s *Server) GetUsersCount(context.Context) int64 {
	return s.validator.GetCount()
}

var _ proxy.UserManager = (*Server)(nil)
