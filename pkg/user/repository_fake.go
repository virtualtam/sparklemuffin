package user

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	// TODO refactor with map[uuid]User
	Users []User
}

func (r *FakeRepository) AddUser(user User) error {
	r.Users = append(r.Users, user)
	return nil
}

func (r *FakeRepository) DeleteUserByUUID(userUUID string) error {
	for index, user := range r.Users {
		if user.UUID == userUUID {
			r.Users = append(r.Users[:index], r.Users[index+1:]...)
		}
	}

	return ErrNotFound
}

func (r *FakeRepository) GetAllUsers() ([]User, error) {
	return r.Users, nil
}

func (r *FakeRepository) GetUserByEmail(email string) (User, error) {
	for _, user := range r.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *FakeRepository) GetUserByRememberTokenHash(rememberTokenHash string) (User, error) {
	for _, user := range r.Users {
		if user.RememberTokenHash == rememberTokenHash {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *FakeRepository) GetUserByUUID(userUUID string) (User, error) {
	for _, user := range r.Users {
		if user.UUID == userUUID {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *FakeRepository) IsUserEmailRegistered(email string) (bool, error) {
	registered := false

	for _, user := range r.Users {
		if user.Email == email {
			registered = true
			break
		}
	}

	return registered, nil
}

func (r *FakeRepository) UpdateUser(user User) error {
	for index, existingUser := range r.Users {
		if existingUser.UUID == user.UUID {
			r.Users[index] = user
			return nil
		}
	}

	return ErrNotFound
}

func (r *FakeRepository) UpdateUserRememberTokenHash(user User) error {
	for index, existingUser := range r.Users {
		if existingUser.UUID == user.UUID {
			r.Users[index].RememberTokenHash = user.RememberTokenHash
			return nil
		}
	}

	return ErrNotFound
}
