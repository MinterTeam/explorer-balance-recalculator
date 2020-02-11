package main

import (
	"github.com/MinterTeam/explorer-balance-recalculator/recalculator"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found")
	}
	rc := recalculator.New()
	rc.Do()

	//nodeApi := api.NewApi(env.NodeApi)
}
