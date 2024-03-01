package internal

import "database/sql"

type User struct {
	Id        uint
	Name      string
	CreatedAt sql.NullTime
}

type Massage struct {
	Id        uint
	Form      User
	To        User
	Content   string
	CreatedAt sql.NullTime
}

type IUserRepository interface {
	Cerate(user *User) error
	Exist(id int) bool
}

type IMassageRepository interface {
	Create(message *Massage) error
	ReadByUserId(userId uint) (*[]Massage, error)
	Delete(ids []uint) error
}
