package queries

import (
	"context"
	"nota/service/db/models"

	"github.com/uptrace/bun"
)

func CreateUser(ctx context.Context, db *bun.DB, email string, name string) (*models.User, error) {
	user := &models.User{
		Email: email,
		Name:  name,
	}
	_, err := db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}
