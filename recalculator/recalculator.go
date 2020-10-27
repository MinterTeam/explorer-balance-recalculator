package recalculator

import (
	"fmt"
	"github.com/MinterTeam/explorer-balance-recalculator/repository"
	"github.com/MinterTeam/minter-explorer-extender/v2/models"
	"github.com/MinterTeam/minter-explorer-tools/v4/helpers"
	"github.com/MinterTeam/minter-go-sdk/v2/api/grpc_client"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"strconv"
	"sync"
	"time"
)

type ReCalculator struct {
	client            *grpc_client.Client
	addressRepository *repository.Address
	balanceRepository *repository.Balance
	blockRepository   *repository.Block
	coinRepository    *repository.Coin
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
		"version": "1.2",
		"app":     "Minter Explorer Balance Re-Calculator",
	})

	nodeApi, err := grpc_client.New(os.Getenv("NODE_API"))
	if err != nil {
		panic(err)
	}

	return &ReCalculator{
		client:            nodeApi,
		addressRepository: repository.NewAddressRepository(),
		balanceRepository: repository.NewBalanceRepository(),
		blockRepository:   repository.NewBlockRepository(),
		coinRepository:    repository.NewCoinRepository(),
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

	rc.logger.Info("Updating balances...")
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

	balancesCount, err := rc.balanceRepository.GetBalancesCount()
	if err != nil {
		rc.logger.WithError(err).Fatal("Saving balances")
	}

	rc.addressRepository.Close()
	rc.balanceRepository.Close()
	rc.blockRepository.Close()
	rc.coinRepository.Close()

	elapsed := time.Since(start)
	rc.logger.Info("Processing time: ", elapsed)
	rc.logger.Info(fmt.Sprintf("%d balances have been handled", balancesCount))
}

func (rc *ReCalculator) GetBalanceFromNode(addresses []*models.Address) ([]*models.Balance, error) {
	var balances []*models.Balance
	list := make([]string, len(addresses))

	for i, address := range addresses {
		a := fmt.Sprintf("Mx%s", address.Address)
		list[i] = a
	}

	balancesFromNode, err := rc.client.Addresses(list)
	if err != nil {
		return nil, err
	}

	for adr, bfn := range balancesFromNode.Addresses {
		addressId, err := rc.addressRepository.FindId(helpers.RemovePrefix(adr))
		if err != nil {
			return nil, err
		}
		for _, b := range bfn.Balance {
			balance := new(models.Balance)
			balance.AddressID = uint(addressId)
			balance.CoinID = uint(b.Coin.Id)
			balance.Value = b.Value
			balances = append(balances, balance)
		}
	}
	return balances, nil
}
