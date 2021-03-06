package repository

import (
	"fmt"
	"github.com/MinterTeam/minter-explorer-extender/v2/models"
	"github.com/go-pg/pg/v10"
	"os"
	"sync"
)

type Address struct {
	db       *pg.DB
	cache    *sync.Map
	invCache *sync.Map
}

func NewAddressRepository() *Address {
	db := pg.Connect(&pg.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		User:     os.Getenv("DB_USER"),
		Database: os.Getenv("DB_NAME"),
		Password: os.Getenv("DB_PASSWORD"),
	})

	return &Address{
		cache:    new(sync.Map),
		invCache: new(sync.Map),
		db:       db,
	}
}

func (r *Address) GetAll() ([]*models.Address, error) {
	var addresses []*models.Address
	err := r.db.Model(&addresses).Select()
	if err == nil {
		r.addToCache(addresses)
	}
	return addresses, err
}

func (r *Address) FindId(address string) (uint, error) {
	//First look in the cache
	id, ok := r.cache.Load(address)
	if ok {
		return id.(uint), nil
	}

	adr := new(models.Address)
	err := r.db.Model(adr).Column("id").Where("address = ?", address).Select(adr)
	if err != nil {
		return 0, err
	}
	return adr.ID, nil
}

func (r *Address) Close() {
	r.db.Close()
}

func (r *Address) addToCache(addresses []*models.Address) {
	for _, a := range addresses {
		_, exist := r.cache.Load(a)
		if !exist {
			r.cache.Store(a.Address, a.ID)
			r.invCache.Store(a.ID, a.Address)
		}
	}
}
