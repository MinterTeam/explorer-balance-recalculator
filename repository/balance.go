package repository

import (
	"fmt"
	"github.com/MinterTeam/minter-explorer-extender/v2/models"
	"github.com/go-pg/pg/v10"
	"os"
)

type Balance struct {
	db *pg.DB
}

func NewBalanceRepository() *Balance {
	return &Balance{
		db: pg.Connect(&pg.Options{
			Addr:     fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
			User:     os.Getenv("DB_USER"),
			Database: os.Getenv("DB_NAME"),
			Password: os.Getenv("DB_PASSWORD"),
		}),
	}
}
func (r *Balance) SaveAll(balances []*models.Balance) error {
	_, err := r.db.Model(&balances).
		OnConflict("(address_id, coin_id) DO UPDATE").
		Insert()
	return err
}

func (r *Balance) GetBalancesCount() (int, error) {
	return r.db.Model((*models.Balance)(nil)).Count()
}

func (r *Balance) Close() {
	r.db.Close()
}
