package main

import (
	"context"
	"github.com/klaytn/klaytn/client"
	"log"
	"math/big"
	"time"
)

func main() {
	start := 6703708 // //4679000   6703708
	end :=  35007488 //35539000 50562000 ~ 50994119 ok

	cli, err := client.Dial("http://52.78.230.244:8551")
	if err != nil {
		log.Println("Dial error")
		return
	}
	//for blockNum := uint64(start); blockNum <= uint64(end); blockNum++ {
	for blockNum := uint64(end); blockNum >= uint64(start); blockNum-- {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		result, err := cli.ValidateBlockHeader(ctx, new(big.Int).SetUint64(blockNum))
		if err != nil {
			log.Println("Found error retry", "blockNum", blockNum, "err", err)
			cancel()
			cli.Close()

			cli, err = client.Dial("http://52.78.230.244:8551")
			if err != nil {
				log.Println("Dial error")
				return
			}

			blockNum++ //blockNum--
			continue
		}

		//if !result.IsValidCommittee || !result.IsValidProposer || !result.IsValidSeal {
		if !result.IsValidProposer || !result.IsValidSeal {
			log.Println("Found error", "blockNum", blockNum,
				"IsValidCommittee", result.IsValidCommittee,
				"IsValidProposer", result.IsValidProposer,
				"IsValidSeal", result.IsValidSeal)
			//return
		}
		cancel()

		if blockNum%1000 == 0 {
			log.Println("validating", "blk", blockNum)
		}
	}
	log.Println("finished validating", "start", start, "end", end)
}
