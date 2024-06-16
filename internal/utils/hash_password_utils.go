package utils

import "golang.org/x/crypto/bcrypt"

type bcryptUtil struct{}

func NewBcrypt() *bcryptUtil {
	return &bcryptUtil{}
}

func (*bcryptUtil) HashPassword(password string) (string, error) {
	const cost = 14
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (*bcryptUtil) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
