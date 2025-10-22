package models

import (
	"time"

	"github.com/uptrace/bun"
)

// --- USERS TABLE ---

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID        int64     `bun:",pk,autoincrement" json:"id"`
	Email     string    `bun:",notnull"         json:"email"`
	Name      string    `bun:",notnull"         json:"name"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relations
	Contents []*Content `bun:"rel:has-many,join:id=user_id" json:"contents,omitempty"`
}

// --- CONTENT TABLE ---

type Content struct {
	bun.BaseModel `bun:"table:content,alias:c"`

	ID        int64     `bun:",pk,autoincrement" json:"id"`
	Title     string    `bun:",notnull"         json:"title"`
	Content   string    `bun:",notnull"         json:"content"`
	UserID    int64     `bun:",notnull"         json:"user_id"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relations
	User *User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}
