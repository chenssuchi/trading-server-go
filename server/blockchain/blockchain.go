package blockchain

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
	"math/big"
	"time"
	trade "trade/binding"
	"trade/blockchain/event"
	"trade/config"
	"trade/db"
	"trade/log"
	"trade/models"
)

type OrderEventType int8

const (
	SaleEvent OrderEventType = iota + 1
	BuyEvent
)

type OrderEvent struct {
	EventType OrderEventType
	OrderId   int64
	Seller    string
	Buyer     string
	Price     string
	Quantity  int64
	SaleAt    int64
	BuyAt     int64
	IsDeal    bool
}

func Client(c config.Conf) *ethclient.Client {
	client, err := ethclient.Dial(c.Blockchain.RPC)
	if err != nil {
		panic("dail failed")
	}
	return client
}

func GetLogs(evt event.Event) {
	Cli := Client(*config.Config)
	err, startHeight, endHeight := GetBlockHeight(Cli)
	if err != nil {
		log.Logger.Sugar().Error("getBlockHeight error:", err)
		return
	}
	query := event.BuildQuery(common.HexToAddress(config.Config.Blockchain.TradeAddress), evt.EventSignature, big.NewInt(startHeight), big.NewInt(endHeight))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
	logs, err := Cli.FilterLogs(ctx, query)
	defer cancel()
	if err != nil {
		log.Logger.Sugar().Error("getLogs error:", err)
		return
	}
	for _, logE := range logs {
		err, orderEvent := ParseSig(logE, evt)
		if err != nil {
			continue
		}
		UpdateOrder(orderEvent)
	}
}

func GetBlockHeight(Cli *ethclient.Client) (error, int64, int64) {
	blockStore := models.BlockStore{}
	number, err := Cli.BlockNumber(context.Background())
	if err != nil {
		return err, 0, 0
	}
	err = db.PG.Table("orders.block_store").Where("id=1").First(&blockStore).Error
	if err == gorm.ErrRecordNotFound {
		db.PG.Table("orders.block_store").Create(&models.BlockStore{
			BlockHeight: number,
		})
	} else {
		return nil, int64(blockStore.BlockHeight), int64(number)
	}
	return nil, 1, int64(number)
}

func ParseSig(logE types.Log, sig event.Event) (error, OrderEvent) {
	orderEvent := OrderEvent{}
	abi, _ := trade.TradeMetaData.GetAbi()
	fmt.Println()
	if sig == event.BuyEvent { // Sale event
		logData, err := abi.Unpack(event.BuyEvent.EventName, logE.Data)
		if err != nil {
			log.Logger.Sugar().Error("get event data error: ", err)
			return err, orderEvent
		}
		orderEvent.EventType = BuyEvent
		orderEvent.OrderId = logE.Topics[1].Big().Int64()
		orderEvent.Seller = logData[0].(common.Address).String()
		orderEvent.Buyer = logData[1].(common.Address).String()
		orderEvent.Price = logData[2].(*big.Int).String()
		orderEvent.Quantity = logData[3].(*big.Int).Int64()
		orderEvent.SaleAt = logData[4].(*big.Int).Int64()
		orderEvent.BuyAt = logData[5].(*big.Int).Int64()
	} else if sig == event.SaleEvent {
		logData, err := abi.Unpack(event.SaleEvent.EventName, logE.Data)
		if err != nil {
			log.Logger.Sugar().Error("get event data error: ", err)
			return err, orderEvent
		}
		orderEvent.EventType = SaleEvent
		orderEvent.OrderId = logE.Topics[1].Big().Int64()
		orderEvent.Seller = logData[0].(common.Address).String()
		orderEvent.Price = logData[1].(*big.Int).String()
		orderEvent.Quantity = logData[2].(*big.Int).Int64()
		orderEvent.SaleAt = logData[3].(*big.Int).Int64()
	}

	return nil, orderEvent
}

func UpdateOrder(orderEvent OrderEvent) {
	orders := models.Orders{}
	SaleAt := time.Unix(orderEvent.SaleAt, 0)
	BuyAt := time.Unix(orderEvent.BuyAt, 0)
	if orderEvent.EventType == SaleEvent {
		whereCondition := fmt.Sprintf("id=%d", orderEvent.OrderId)
		err := db.PG.Table("orders.orders").Where(whereCondition).First(&orders).Error
		if err == gorm.ErrRecordNotFound {
			err = db.PG.Table("orders.orders").Create(&models.Orders{
				OrderId:  orderEvent.OrderId,
				Seller:   orderEvent.Seller,
				Buyer:    orderEvent.Buyer,
				Price:    orderEvent.Price,
				Quantity: orderEvent.Quantity,
				SaleAt:   SaleAt.Format("2006-01-02 15:04:05"),
				BuyAt:    BuyAt.Format("2006-01-02 15:04:05"),
				IsDeal:   orderEvent.IsDeal,
			}).Error
		}
	} else if orderEvent.EventType == BuyEvent {
		whereCondition := fmt.Sprintf("id=%d", orderEvent.OrderId)
		err := db.PG.Table("orders.orders").Where(whereCondition).First(&orders).Error
		if err != gorm.ErrRecordNotFound && orders.IsDeal == false {
			err = db.PG.Table("orders.orders").Updates(&models.Orders{
				OrderId:  orderEvent.OrderId,
				Seller:   orderEvent.Seller,
				Buyer:    orderEvent.Buyer,
				Price:    orderEvent.Price,
				Quantity: orderEvent.Quantity,
				SaleAt:   SaleAt.Format("2006-01-02 15:04:05"),
				BuyAt:    BuyAt.Format("2006-01-02 15:04:05"),
				IsDeal:   true,
			}).Error
		}
	}
}

func GetAuth(cli *ethclient.Client) (error, *bind.TransactOpts) {
	privateKeyEcdsa, err := crypto.HexToECDSA(config.Config.Blockchain.PrivateKey)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return err, nil
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKeyEcdsa, big.NewInt(int64(config.Config.Blockchain.ChainId)))
	if err != nil {
		log.Logger.Sugar().Error(err)
		return err, nil
	}

	return nil, &bind.TransactOpts{
		From:      auth.From,
		Nonce:     nil,
		Signer:    auth.Signer, // Method to use for signing the transaction (mandatory)
		Value:     big.NewInt(0),
		GasPrice:  nil,
		GasFeeCap: nil,
		GasTipCap: nil,
		GasLimit:  0,
		Context:   context.Background(),
		NoSend:    false, // Do all transact steps but do not send the transaction
	}
}

func Approve(cli *ethclient.Client) {
	err, txOps := GetAuth(cli)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
	IntoToken, err := trade.NewIntoToken(common.HexToAddress(config.Config.Blockchain.IntoTokenAddress), cli)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
	amountAdd := big.NewInt(100000000000000000)
	amountRes := big.NewInt(100000000000000000).Mul(amountAdd, amountAdd)
	tx, err := IntoToken.Approve(txOps, common.HexToAddress(config.Config.Blockchain.TradeAddress), amountRes)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
	log.Logger.Sugar().Info("approve ok ", tx.Hash())
}

func AdminSetTokenAddress(cli *ethclient.Client) {
	err, txOps := GetAuth(cli)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
	IntoTrade, err := trade.NewTrade(common.HexToAddress(config.Config.Blockchain.TradeAddress), cli)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
	tx, err := IntoTrade.AdminSetTokenAddress(txOps, common.HexToAddress(config.Config.Blockchain.IntoTokenAddress))
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
	log.Logger.Sugar().Info("adminSetTokenAddress ok ", tx.Hash())
}

func Sale(cli *ethclient.Client) {
	err, txOps := GetAuth(cli)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
	IntoTrade, err := trade.NewTrade(common.HexToAddress(config.Config.Blockchain.TradeAddress), cli)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
	amount := big.NewInt(1000000000000000000)
	tx, err := IntoTrade.Sale(txOps, amount, amount)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
	log.Logger.Sugar().Info("sale ok ", tx.Hash())
}

func Buy(cli *ethclient.Client) {
	err, txOps := GetAuth(cli)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
	IntoTrade, err := trade.NewTrade(common.HexToAddress(config.Config.Blockchain.TradeAddress), cli)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
	orderId := big.NewInt(1)
	log.Logger.Sugar().Info(config.Config.Blockchain.TradeAddress)
	tx, err := IntoTrade.Buy(txOps, orderId)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
	log.Logger.Sugar().Info("buy ok ", tx.Hash())
}
