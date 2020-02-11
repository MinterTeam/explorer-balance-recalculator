<p align="center" background="black"><img src="minter-logo.svg" width="400"></p>

<p align="center" style="text-align: center;">
    <a href="https://github.com/MinterTeam/explorer-genesis-uploader/blob/master/LICENSE">
        <img src="https://img.shields.io/packagist/l/doctrine/orm.svg" alt="License">
    </a>
    <img alt="undefined" src="https://img.shields.io/github/last-commit/MinterTeam/explorer-genesis-uploader.svg">
</p>

# Minter Explorer Balance Re-Calculator

The official repository of Minter Explorer Balance Re-Calculator service.

Minter Explorer Balance Re-Calculator is a service which provides to update balances in database Minter Explorer.

## Requirement

- PostgresSQL

## Build

- run `go mod tidy`

- run `go build -o ./builds/recalculator ./cmd/recalculator.go`

## Run

- copy `.env.prod` to `.env` and fill with own values

- run `./builds/recalculator` or `docker-compose up`