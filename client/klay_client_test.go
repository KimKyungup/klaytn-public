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
	"bytes"
	"context"
	"encoding/json"
	"github.com/klaytn/klaytn"
	"github.com/klaytn/klaytn/blockchain/types"
	"github.com/klaytn/klaytn/common"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/big"
	"net/http"
	"testing"
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

func Test_ExampleKASClient(t *testing.T) {
	header := map[string]string{
		"x-krn": "krn:1001:node",
	}
	id := "78ab9116689659321aaf472aa154eac7dd7a99c6"
	pass := "403e0397d51a823cd59b7edcb212788c8599dd7e"

	c, err := DialWithHeader("https://node-api.beta.klaytn.io/v1/klaytn", header, id, pass)
	assert.NoError(t, err)

	ctx := context.Background()
	blkNum, err := c.BlockNumber(ctx)
	assert.NoError(t, err)

	t.Log(blkNum.String())
}

func Test_ExampleKASClient2(t *testing.T) {
	operator := common.StringToAddress("0x1552F52D459B713E0C4558e66C8c773a75615FA8")

	user := "78ab9116689659321aaf472aa154eac7dd7a99c6"
	pwd := "403e0397d51a823cd59b7edcb212788c8599dd7e"

	// HTTP Request Body
	type Body struct {
		Operator common.Address `json:"operator"`
		Payload  interface{} `json:"payload"`
	}

	// API Payload
	type Payload struct {
		Id string `json:"id"`
		types.AnchoringDataInternalType0
	}

	// Anchor Data
	AnchorData := &types.AnchoringDataInternalType0{
		BlockHash:     common.HexToHash("0"),
		TxHash:        common.HexToHash("1"),
		ParentHash:    common.HexToHash("2"),
		ReceiptHash:   common.HexToHash("3"),
		StateRootHash: common.HexToHash("4"),
		BlockNumber:   big.NewInt(5),
		BlockCount:    big.NewInt(6),
		TxCount:       big.NewInt(7),
	}

	payload := &Payload{
		Id: AnchorData.BlockNumber.String(),
		AnchoringDataInternalType0: *AnchorData,
	}

	bodyData := Body{
		Operator: operator,
		Payload: payload,
	}
	payloadBytes, err := json.Marshal(bodyData)
	if err != nil {
		// handle err
	}
	t.Log(string(payloadBytes))
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "http://anchor.wallet-api.dev.klaytn.com/v1/anchor", body)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth(user, pwd)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Krn", "krn:1001:anchor:test:operator-pool:op1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	resBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(resBody))
	var v interface{}
	err = json.Unmarshal(resBody, &v)
	if err != nil {
		t.Fatal(err)
	}

}
