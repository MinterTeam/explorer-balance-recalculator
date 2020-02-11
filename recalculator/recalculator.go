package recalculator

import (
	"fmt"
	"github.com/MinterTeam/explorer-balance-recalculator/repository"
	"github.com/MinterTeam/minter-explorer-tools/v4/helpers"
	"github.com/MinterTeam/minter-explorer-tools/v4/models"
	"github.com/MinterTeam/minter-go-sdk/api"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"strconv"
	"sync"
	"time"
)

type ReCalculator struct {
	client            *api.Api
	addressRepository *repository.Address
	coinRepository    *repository.Coin
	balanceRepository *repository.Balance
	logger            *logrus.Entry
}

func New() *ReCalculator {
	//Init Logger
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetReportCaller(false)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	contextLogger := logger.WithFields(logrus.Fields{
		"version": "1.0",
		"app":     "Minter Explorer Balance Re-Calculator",
	})

	return &ReCalculator{
		client:            api.NewApi(os.Getenv("NODE_API")),
		addressRepository: repository.NewAddressRepository(),
		coinRepository:    repository.NewCoinRepository(),
		balanceRepository: repository.NewBalanceRepository(),
		logger:            contextLogger,
	}
}

func (rc *ReCalculator) Do() {
	start := time.Now()

	rc.logger.Info("Getting addresses from DB...")
	addresses, err := rc.addressRepository.GetAll()
	if err != nil {
		rc.logger.WithError(err).Fatal("Getting addresses from DB")
	}

	var balances []*models.Balance
	ch := make(chan []*models.Balance)
	go func() {
		for b := range ch {
			balances = append(balances, b...)
		}
	}()

	wgAddresses := new(sync.WaitGroup)
	addrChunkSize := os.Getenv("APP_ADDRESS_CHUNK_SIZE")
	chunkSize, err := strconv.ParseInt(addrChunkSize, 10, 64)
	if err != nil {
		rc.logger.WithError(err).Fatal("Getting balances")
	}
	chunksCount := int(math.Ceil(float64(len(addresses)) / float64(chunkSize)))
	rc.logger.Info("Getting balances from Node...")
	for i := 0; i < chunksCount; i++ {
		start := int(chunkSize) * i
		end := start + int(chunkSize)
		if end > len(addresses) {
			end = len(addresses)
		}
		wgAddresses.Add(1)
		go func() {
			b, err := rc.GetBalanceFromNode(addresses[start:end])
			if err != nil {
				rc.logger.WithError(err).Fatal("Getting balances from Node")
				wgAddresses.Done()
			}
			ch <- b
			wgAddresses.Done()
		}()
	}
	wgAddresses.Wait()
	close(ch)

	rc.logger.Info("Saving balances from Node...")
	wgBalances := new(sync.WaitGroup)
	balanceChunkSize := os.Getenv("APP_BALANCES_CHUNK_SIZE")
	chunkSize, err = strconv.ParseInt(balanceChunkSize, 10, 64)
	if err != nil {
		rc.logger.WithError(err).Fatal("Saving balances")
	}
	chunksCount = int(math.Ceil(float64(len(balances)) / float64(chunkSize)))

	for i := 0; i < chunksCount; i++ {
		start := int(chunkSize) * i
		end := start + int(chunkSize)
		if end > len(balances) {
			end = len(balances)
		}
		wgBalances.Add(1)
		balancesModels := balances[start:end]
		go func(balances []*models.Balance) {
			err = rc.balanceRepository.SaveAll(balances)
			if err != nil {
				rc.logger.WithError(err).Fatal("Saving balances")
			}
			wgBalances.Done()
		}(balancesModels)
	}
	wgBalances.Wait()

	elapsed := time.Since(start)
	rc.logger.Info("Processing time: ", elapsed)
}

func (rc *ReCalculator) GetBalanceFromNode(addresses []*models.Address) ([]*models.Balance, error) {
	var balances []*models.Balance
	list := make([]string, len(addresses))

	for i, address := range addresses {
		a := fmt.Sprintf("\"Mx%s\"", address.Address)
		list[i] = a
	}

	currentBlock, err := strconv.ParseInt(os.Getenv("CURRENT_BLOCK"), 10, 64)
	if err != nil {
		return nil, err
	}

	balancesFromNode, err := rc.client.Addresses(list, int(currentBlock))
	if err != nil {
		return nil, err
	}

	for _, bfn := range balancesFromNode {
		addressId, err := rc.addressRepository.FindId(helpers.RemovePrefix(bfn.Address))
		if err != nil {
			return nil, err
		}
		for coin, value := range bfn.Balance {
			if value == "0" && coin != os.Getenv("APP_BASE_COIN") {
				continue
			}
			coinId, err := rc.coinRepository.FindIdBySymbol(coin)
			if err != nil {
				return nil, err
			}
			balance := new(models.Balance)
			balance.AddressID = addressId
			balance.CoinID = coinId
			balance.Value = value
			balances = append(balances, balance)
		}
	}
	return balances, nil
}
