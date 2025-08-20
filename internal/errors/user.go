package errors

import "errors"

var (
	// Возвращается, если пользователь с таким username уже существует.
	ErrUserAlreadyExists = errors.New("user already exists")

	// Возвращается, если пользователь с заданным UUID не найден.
	ErrUserNotFound = errors.New("user not found")
)
