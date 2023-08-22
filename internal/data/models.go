package data

import "database/sql"

type Models struct {
	Users   UserModel
	Lichess LichessModel
}

func NewModels(db *sql.DB) Models {

	return Models{
		Users:   UserModel{DB: db},
		Lichess: LichessModel{DB: db},
	}

}
