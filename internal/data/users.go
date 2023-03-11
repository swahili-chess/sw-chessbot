package data

import (
	"context"
	"database/sql"
	"time"
)

type User struct {
	ID       int64
	Isactive bool
}

type UserModel struct {
	DB *sql.DB
}

type UserId struct {
	Id int64
}

func (u UserModel) Insert(user *User) error {

	query := `INSERT INTO users(id, isactive) VALUES ($1, $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	_, err := u.DB.ExecContext(ctx, query, user.ID, user.Isactive)

	return err

}

func (u UserModel) Update(user *User) error {
	query := `UPDATE users SET isactive = $1 WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	_, err := u.DB.ExecContext(ctx, query, user.Isactive, user.ID)

	return err
}

func (u UserModel) GetActiveUsers() ([]UserId, error) {
	query := `SELECT id from users WHERE isactive = true`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	rows, err := u.DB.QueryContext(ctx,query)

	if err != nil {
		return nil, err
	}

	var userIds []UserId

	for rows.Next() {
        var userid UserId
		if err:= rows.Scan(&userid.Id); err!= nil {
           return userIds, err
		}

		userIds = append(userIds, userid)
	}

	if err := rows.Err(); err != nil {
	    return userIds, err	
	}
	return userIds, nil
}
