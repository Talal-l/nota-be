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

type CreateContentArgs struct {
	Content string
	UserID  int64
}

func CreateContent(ctx context.Context, db *bun.DB, args CreateContentArgs) (*models.Content, error) {
	content := &models.Content{
		Content: args.Content,
		UserID:  args.UserID,
	}
	_, err := db.NewInsert().Model(content).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func GetContents(ctx context.Context, db *bun.DB) ([]models.Content, error) {
	var content []models.Content
	err := db.NewSelect().Model(&content).Relation("User").Scan(ctx)
	if err != nil {
		return nil, err
	}

	return content, nil
}
