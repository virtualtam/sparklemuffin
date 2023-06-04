package bookmark

func requireUserUUID(userUUID string) error {
	if userUUID == "" {
		return ErrUserUUIDRequired
	}
	return nil
}
