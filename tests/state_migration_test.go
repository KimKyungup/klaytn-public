package tests

import (
	"crypto/ecdsa"
	"github.com/klaytn/klaytn/blockchain"
	"github.com/klaytn/klaytn/blockchain/types"
	"github.com/klaytn/klaytn/common"
	"github.com/klaytn/klaytn/networks/p2p"
	"github.com/klaytn/klaytn/node"
	"github.com/klaytn/klaytn/node/cn"
	"github.com/klaytn/klaytn/params"
	"github.com/klaytn/klaytn/ser/rlp"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"
)

func testKlaytnNode(t *testing.T, dir string, validator *TestAccountType) (*node.Node, *cn.CN) {
	var klaytnNode *cn.CN

	fullNode, err := node.New(&node.Config{DataDir: dir, UseLightweightKDF: true, P2P: p2p.Config{PrivateKey: validator.Keys[0]}})
	if err != nil {
		t.Fatalf("failed to create node: %v", err)
	}

	istanbulConfData, err := rlp.EncodeToBytes(&types.IstanbulExtra{
		Validators:    []common.Address{validator.Addr},
		Seal:          []byte{},
		CommittedSeal: [][]byte{},
	})
	if err != nil {
		t.Fatal(err)
	}

	genesis := blockchain.DefaultGenesisBlock()
	genesis.ExtraData = genesis.ExtraData[:types.IstanbulExtraVanity]
	genesis.ExtraData = append(genesis.ExtraData, istanbulConfData...)
	genesis.Config.Istanbul.SubGroupSize = 1
	genesis.Config.Governance.Reward.MintingAmount = new(big.Int).Mul(big.NewInt(100), big.NewInt(params.KLAY))

	cnConf := &cn.DefaultConfig
	cnConf.Genesis = genesis
	cnConf.Rewardbase = validator.Addr
	cnConf.PartitionedDB = true

	if err = fullNode.Register(func(ctx *node.ServiceContext) (node.Service, error) { return cn.New(ctx, cnConf) }); err != nil {
		t.Fatalf("failed to register Klaytn protocol: %v", err)
	}

	if err = fullNode.Start(); err != nil {
		t.Fatalf("failed to start test fullNode: %v", err)
	}

	if err := fullNode.Service(&klaytnNode); err != nil {
		t.Fatal(err)
	}

	return fullNode, klaytnNode
}

func TestStateMigration(t *testing.T) {
	if testing.Verbose() {
		enableLog() // Change verbosity level in the function if you need
	}

	// Prepare workspace
	workspace, err := ioutil.TempDir("", "klaytn-test-state")
	if err != nil {
		t.Fatalf("failed to create temporary keystore: %v", err)
	}
	defer os.RemoveAll(workspace)

	// Prepare a validator
	validator, err := createAnonymousAccount(getRandomPrivateKeyString(t))
	if err != nil {
		t.Fatal(err)
	}

	// Create a Klaytn node
	fullNode, node := testKlaytnNode(t, workspace, validator)
	if err := node.StartMining(false); err != nil {
		t.Fatal()
	}
	time.Sleep(2 * time.Second) // wait for initializing mining

	// Accounts used in the test
	const numAccounts = 3
	var accounts [numAccounts]*TestAccountType

	richAccount := &TestAccountType{
		Addr:  validator.Addr,
		Keys:  []*ecdsa.PrivateKey{validator.Keys[0]},
		Nonce: uint64(0),
	}

	for i := 0; i < numAccounts; i++ {
		if accounts[i], err = createAnonymousAccount(getRandomPrivateKeyString(t)); err != nil {
			t.Fatal()
		}
	}

	// Variables used for tx generation
	var tx *types.Transaction
	chainId := node.BlockChain().Config().ChainID
	signer := types.NewEIP155Signer(chainId)
	gasPrice := big.NewInt(25 * params.Ston)
	gasLimit = uint64(100000)

	// Generate a transaction and add to the txpool
	tx, _ = genLegacyTransaction(t, signer, richAccount, accounts[0], nil, gasPrice)
	if err := node.TxPool().AddLocal(tx); err != nil {
		t.Fatal(err)
	}
	richAccount.AddNonce()
	time.Sleep(2 * time.Second)

	time.Sleep(blockchain.DefaultBlockInterval * time.Second)

	// Prepare a state migration block number
	currentBlock := node.BlockChain().CurrentBlock()
	migrationBlockNum := currentBlock.NumberU64() - (currentBlock.NumberU64() % blockchain.DefaultBlockInterval)
	migrationBlock := node.BlockChain().GetBlockByNumber(migrationBlockNum)

	// Start state migration
	if err := node.BlockChain().StartStateMigration(migrationBlockNum, migrationBlock.Root()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second) // give a time for state migration

	// Stop full node
	if err := fullNode.Stop(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)

	// Start full node with previous db
	fullNode, node = testKlaytnNode(t, workspace, validator)
	if err := node.StartMining(false); err != nil {
		t.Fatal()
	}
	if err := node.StartMining(false); err != nil {
		t.Fatal()
	}
	time.Sleep(20 * time.Second)
}
