package models

import "time"

type Role string

const (
	RoleEmployee  Role = "employee"
	RoleModerator Role = "moderator"
)

type User struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password,omitempty" json:"-"`
	Role      Role      `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
