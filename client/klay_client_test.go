// Modifications Copyright 2018 The klaytn Authors
// Copyright 2016 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from ethclient/ethclient_test.go (2018/06/04).
// Modified and improved for the klaytn development.

package client

import (
	"context"
	"github.com/klaytn/klaytn"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

// Verify that Client implements the Klaytn interfaces.
var (
	// _ = klaytn.Subscription(&Client{})
	_ = klaytn.ChainReader(&Client{})
	_ = klaytn.TransactionReader(&Client{})
	_ = klaytn.ChainStateReader(&Client{})
	_ = klaytn.ChainSyncReader(&Client{})
	_ = klaytn.ContractCaller(&Client{})
	_ = klaytn.LogFilterer(&Client{})
	_ = klaytn.TransactionSender(&Client{})
	_ = klaytn.GasPricer(&Client{})
	_ = klaytn.PendingStateReader(&Client{})
	_ = klaytn.PendingContractCaller(&Client{})
	_ = klaytn.GasEstimator(&Client{})
	// _ = klaytn.PendingStateEventer(&Client{})
)

func Benchmark_GetBlockNumber(b *testing.B) {
	cli, err := Dial("http://localhost:8551")
	if err != nil {
		b.Fatal("Dial error")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	blkNum, err := cli.BlockNumber(ctx)
	if err != nil {
		b.Fatal("BlockNumber error")
	}

	rand.Seed(time.Now().UnixNano())

	successCnt := 0
	failCnt := 0
	b.ResetTimer()
	for k := 0; k < b.N; k++ {
		successCnt++
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		_, err := cli.BlockByNumber(ctx, new(big.Int).SetUint64(rand.Uint64()%blkNum.Uint64()))
		if err != nil {
			failCnt++
			b.Log("fail", "failCnt", failCnt, "err", err)
		}
		cancel()
	}

	b.Log("report", "successCnt", successCnt, "failCnt", failCnt)
}

//
//func Test_GetBlockNumber(b *testing.T) {
//	cli, err := Dial("http://localhost:8551")
//	if err != nil {
//		b.Fatal("Dial error")
//	}
//
//	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//	//defer cancel()
//	//blkNum, err := cli.BlockNumber(ctx)
//	//if err != nil {
//	//	b.Fatal("BlockNumber error")
//	//}
//
//	rand.Seed(time.Now().UnixNano())
//
//	successCnt := 0
//	failCnt := 0
//
//	var chan
//	for i := 0; i < 10; i++ {
//		go
//	}
//
//
//	start := time.Now()
//	for k := uint64(0); k < 100; k++ {
//		successCnt++
//		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//
//		_, err := cli.BlockByNumber(ctx, new(big.Int).SetUint64(k)) //rand.Uint64()%blkNum.Uint64()
//		if err != nil {
//			failCnt++
//			b.Log("fail", "failCnt", failCnt, "err", err)
//		}
//		cancel()
//	}
//
//	b.Log("report", "successCnt", successCnt, "failCnt", failCnt, "elapsed", time.Since(start))
//}

func Test_GetBlockNumber(b *testing.T) {
	cli, err := Dial("http://localhost:8551")
	if err != nil {
		b.Fatal("Dial error")
	}

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//lastBlkNum, err := cli.BlockNumber(ctx)
	//if err != nil {
	//	b.Fatal("BlockNumber error")
	//}

	for i := uint64(10000); i < 20000; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		result, err := cli.ValidateBlockHeader(ctx, new(big.Int).SetUint64(i))
		if err != nil {
			b.Log("Found error", "blockNum", i, "err", err)
			i--
			cancel()
			continue
		}

		if !result.IsValidCommittee || !result.IsValidProposer || !result.IsValidSeal {
			b.Log("Found error", "blockNum", i,
				"IsValidCommittee", result.IsValidCommittee,
				"IsValidProposer", result.IsValidProposer,
				"IsValidSeal", result.IsValidSeal)
		}
		cancel()
		//b.Log("OK", "blockNum",i)
	}
}
