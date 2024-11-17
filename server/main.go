package main

import (
	"trade/blockchain"
	"trade/db"
)

func main() {
	db.InitPg()
	//Cli := blockchain.Client(*config.Config)
	//blockchain.Approve(Cli)
	//blockchain.AdminSetTokenAddress(Cli)
	//blockchain.Sale(Cli)
	//blockchain.Buy(Cli)
	blockchain.GetBlockHeight()
	//blockchain.GetLogs(Cli, event.SaleEvent, 39258755, 39258755)
}
