package tasks

import (
	"github.com/jasonlvhit/gocron"
	"time"
	"trade/blockchain"
	"trade/blockchain/event"
)

func Task() {

	s := gocron.NewScheduler()
	s.ChangeLoc(time.UTC)
	_ = s.Every(5).Seconds().From(gocron.NextTick()).Do(blockchain.GetLogs, event.SaleEvent)
	<-s.Start() // Start all the pending jobs

}
