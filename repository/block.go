package repository

import (
	"fmt"
	"github.com/MinterTeam/minter-explorer-extender/v2/models"
	"github.com/go-pg/pg/v10"
	"os"
)

type Block struct {
	db *pg.DB
}

func NewBlockRepository() *Block {
	return &Block{
		db: pg.Connect(&pg.Options{
			Addr:     fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
			User:     os.Getenv("DB_USER"),
			Database: os.Getenv("DB_NAME"),
			Password: os.Getenv("DB_PASSWORD"),
		}),
	}
}

func (r *Block) GetLastBlockId() (int, error) {
	var id int
	err := r.db.Model((*models.Block)(nil)).
		Column("id").
		Order("id DESC").
		Limit(1).
		Select(&id)

	return id, err
}

func (r *Block) Close() {
	r.db.Close()
}
