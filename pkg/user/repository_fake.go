package user

var _ Repository = &fakeRepository{}

type fakeRepository struct {
	users []User
}

func (r *fakeRepository) AddUser(user User) error {
	r.users = append(r.users, user)
	return nil
}

func (r *fakeRepository) GetUserByEmail(email string) (User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}
