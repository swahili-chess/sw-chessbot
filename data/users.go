package data

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type User struct {
	ID       string
	Isactive string
}

type UserModel struct {
	DB *sql.DB
}

func (u UserModel) Insert(user *User) error {

	query := `INSERT INTO users(id, isactive) VALUES ($1, $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()


	_ , err := u.DB.ExecContext(ctx,query,user.ID,user.Isactive)

	if err != nil {
		log.Println(err)
	}
    
	return err
	

}
