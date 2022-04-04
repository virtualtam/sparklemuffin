package user

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	// TODO refactor with map[uuid]User
	Users []User
}

func (r *FakeRepository) UserAdd(user User) error {
	r.Users = append(r.Users, user)
	return nil
}

func (r *FakeRepository) UserDeleteByUUID(userUUID string) error {
	for index, user := range r.Users {
		if user.UUID == userUUID {
			r.Users = append(r.Users[:index], r.Users[index+1:]...)
		}
	}

	return ErrNotFound
}

func (r *FakeRepository) UserGetAll() ([]User, error) {
	return r.Users, nil
}

func (r *FakeRepository) UserGetByEmail(email string) (User, error) {
	for _, user := range r.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *FakeRepository) UserGetByNickName(nick string) (User, error) {
	for _, user := range r.Users {
		if user.NickName == nick {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *FakeRepository) UserGetByUUID(userUUID string) (User, error) {
	for _, user := range r.Users {
		if user.UUID == userUUID {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *FakeRepository) UserIsEmailRegistered(email string) (bool, error) {
	registered := false

	for _, user := range r.Users {
		if user.Email == email {
			registered = true
			break
		}
	}

	return registered, nil
}

func (r *FakeRepository) UserIsNickNameRegistered(nick string) (bool, error) {
	registered := false

	for _, user := range r.Users {
		if user.NickName == nick {
			registered = true
			break
		}
	}

	return registered, nil
}

func (r *FakeRepository) UserUpdate(user User) error {
	for index, existingUser := range r.Users {
		if existingUser.UUID == user.UUID {
			r.Users[index] = user
			return nil
		}
	}

	return ErrNotFound
}

func (r *FakeRepository) UserUpdateInfo(info InfoUpdate) error {
	for index, existingUser := range r.Users {
		if existingUser.UUID == info.UUID {
			r.Users[index].Email = info.Email
			r.Users[index].UpdatedAt = info.UpdatedAt
			return nil
		}
	}

	return ErrNotFound
}

func (r *FakeRepository) UserUpdatePasswordHash(passwordHash PasswordHashUpdate) error {
	for index, existingUser := range r.Users {
		if existingUser.UUID == passwordHash.UUID {
			r.Users[index].PasswordHash = passwordHash.PasswordHash
			r.Users[index].UpdatedAt = passwordHash.UpdatedAt
			return nil
		}
	}

	return ErrNotFound
}
