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

func (r *FakeRepository) UpdateUser(user User) error {
	for index, existingUser := range r.Users {
		if existingUser.UUID == user.UUID {
			r.Users[index] = user
			return nil
		}
	}

	return ErrNotFound
}