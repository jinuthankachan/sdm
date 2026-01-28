package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Save(ctx context.Context, model *User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		pii := UserPii{
			Id:      model.Id,
			Ssn:     model.Ssn,
			Address: model.Address,
		}
		if err := tx.Create(&pii).Error; err != nil {
			return err
		}

		// Save Chain Fields
		if err := tx.Create(&UserChain{
			Key:        model.Id,
			FieldName:  "id",
			FieldValue: fmt.Sprintf("%v", model.Id),
		}).Error; err != nil {
			return err
		}
		// Hash Address
		h_Address := sha256.Sum256([]byte(fmt.Sprintf("%v", model.Address)))
		hashed_Address := hex.EncodeToString(h_Address[:])
		if err := tx.Create(&UserChain{
			Key:        model.Id,
			FieldName:  "hashed_address",
			FieldValue: hashed_Address,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&UserChain{
			Key:        model.Id,
			FieldName:  "name",
			FieldValue: fmt.Sprintf("%v", model.Name),
		}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *UserRepo) Fetch(ctx context.Context, id string) (*UserView, error) {
	var view UserView
	// GORM might not support querying Views directly with First if it doesn't know it's a table.
	// But we defined TableName() to return the view name, so it should work.
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&view).Error; err != nil {
		return nil, err
	}
	return &view, nil
}
