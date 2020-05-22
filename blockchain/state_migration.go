// Copyright 2019 The klaytn Authors
// This file is part of the klaytn library.
//
// The klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the klaytn library. If not, see <http://www.gnu.org/licenses/>.

package blockchain

import (
	"errors"
	"fmt"
	"github.com/klaytn/klaytn/blockchain/state"
	"github.com/klaytn/klaytn/common"
	"github.com/klaytn/klaytn/storage/database"
	"github.com/klaytn/klaytn/storage/statedb"
	"math"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type stateTrieMigrationDB struct {
	database.DBManager
}

func (td *stateTrieMigrationDB) ReadCachedTrieNode(hash common.Hash) ([]byte, error) {
	return td.ReadCachedTrieNodeFromNew(hash)
}
func (td *stateTrieMigrationDB) ReadCachedTrieNodePreimage(secureKey []byte) ([]byte, error) {
	return td.ReadCachedTrieNodePreimageFromNew(secureKey)
}

func (td *stateTrieMigrationDB) ReadStateTrieNode(key []byte) ([]byte, error) {
	return td.ReadStateTrieNodeFromNew(key)
}

func (td *stateTrieMigrationDB) HasStateTrieNode(key []byte) (bool, error) {
	return td.HasStateTrieNodeFromNew(key)
}

func (td *stateTrieMigrationDB) ReadPreimage(hash common.Hash) []byte {
	return td.ReadPreimageFromNew(hash)
}

func (bc *BlockChain) stateMigrationCommit(s *statedb.TrieSync, db database.DBManager) (int, time.Duration, error) {
	start := time.Now()
	stateTrieBatch := db.NewBatch(database.StateTrieDB)

	written, err := s.Commit(stateTrieBatch)
	if written == 0 || err != nil {
		return written, 0, err
	}

	if err := stateTrieBatch.Write(); err != nil {
		return 0, 0, fmt.Errorf("DB write error: %v", err)
	}

	return written, time.Since(start), nil
}

func (bc *BlockChain) concurrentRead(db *statedb.Database, quitCh chan struct{}, hashCh chan common.Hash, resultCh chan statedb.SyncResult) {
	for {
		select {
		case <-quitCh:
			return
		case hash := <-hashCh:
			data, err := db.NodeFromOld(hash)
			if err != nil {
				resultCh <- statedb.SyncResult{Hash: hash, Err: err}
				continue
			}
			resultCh <- statedb.SyncResult{Hash: hash, Data: data}
		}
	}
}

func (bc *BlockChain) migrateState(rootHash common.Hash) error {
	bc.wg.Add(1)
	defer bc.wg.Done()

	start := time.Now()
	stateDB := bc.StateCache()
	trieDB := stateDB.TrieDB()

	newWrappedDiskDB := &stateTrieMigrationDB{bc.db}
	newStateDB := state.NewDatabaseWithCache(newWrappedDiskDB, 512, 0)
	newTrieDB := newStateDB.TrieDB()

	trieSync := state.NewStateSync(rootHash, newWrappedDiskDB)
	var queue []common.Hash
	committedCnt := 0

	quitCh := make(chan struct{})
	defer close(quitCh)

	// Prepare concurrent read goroutines
	threads := int(math.Max(math.Ceil(float64(runtime.NumCPU())/2), 4))
	hashCh := make(chan common.Hash, threads)
	resultCh := make(chan statedb.SyncResult, threads)

	for th := 0; th < threads; th++ {
		go bc.concurrentRead(trieDB, quitCh, hashCh, resultCh)
	}

	// Migration main loop
	for trieSync.Pending() > 0 {
		bc.committedCnt, bc.pendingCnt = committedCnt, trieSync.Pending()
		queue = append(queue[:0], trieSync.Missing(database.IdealBatchSize)...)
		results := make([]statedb.SyncResult, len(queue))

		// Read the trie nodes
		startIter := time.Now()
		go func() {
			for _, hash := range queue {
				hashCh <- hash
			}
		}()

		for i := 0; i < len(queue); i++ {
			result := <-resultCh
			if result.Err != nil {
				return fmt.Errorf("failed to retrieve node data for %x: %v", result.Hash, result.Err)
			}
			results[i] = result
		}
		read, readElapsed := len(queue), time.Since(startIter)

		// Process trie nodes
		if _, index, err := trieSync.Process(results); err != nil {
			return fmt.Errorf("failed to process result #%d: %v", index, err)
		}

		// Commit trie nodes
		written, writeElapsed, err := bc.stateMigrationCommit(trieSync, newTrieDB.DiskDB())
		if err != nil {
			return fmt.Errorf("failed to commit data #%d: %v", written, err)
		}

		// Report progress
		committedCnt += written
		bc.committedCnt, bc.pendingCnt, bc.progress = committedCnt, trieSync.Pending(), trieSync.CalcProgressPercentage()
		progressStr := strconv.FormatFloat(bc.progress, 'f', 4, 64)
		progressStr = strings.TrimRight(progressStr, "0")
		progressStr = strings.TrimRight(progressStr, ".") + "%"

		logger.Warn("State migration progress",
			"progress", progressStr, "committedCnt", committedCnt, "pendingCnt", bc.pendingCnt,
			"read", read, "readElapsed", readElapsed, "written", written, "writeElapsed", writeElapsed,
			"elapsed", time.Since(startIter))

		select {
		case <-bc.stopStateMigration:
			// TODO-Klaytn Revert DB.
			// - copied new DB data to old DB.
			// - remove new DB
			return errors.New("stop state migration")
		case <-bc.quit:
			return nil
		default:
		}
	}
	bc.committedCnt, bc.pendingCnt, bc.progress = committedCnt, trieSync.Pending(), trieSync.CalcProgressPercentage()

	elapsed := time.Since(start)
	speed := float64(committedCnt) / elapsed.Seconds()
	logger.Info("State migration is completed", "committedCnt", committedCnt, "elapsed", elapsed, "committed per second", speed)

	// Preimage Copy
	// TODO-Klaytn consider to copy preimage

	// Cross check that the two tries are in sync
	// TODO-Klaytn consider to check Trie contents optionally
	// TODO-Klaytn consider to check storage trie also
	//dirty, err := bc.checkTrieContents(targetDB, srcCachedDB, rootHash)
	//if err != nil || len(dirty) > 0 {
	//	logger.Error("copied state is invalid", "err", err, "len(dirty)", len(dirty))
	//	// TODO-Klaytn Remove new DB and log.Error
	//	if err != nil {
	//		return err
	//	}
	//
	//	return errors.New("copied state is not same with origin")
	//}

	err := CompareStateConsistency(stateDB, newStateDB, rootHash)
	if err != nil {
		logger.Error("State migration failed", "err", err)
		return err
	}

	bc.db.FinishStateMigration()
	logger.Info("completed state migration")

	return nil
}

func (bc *BlockChain) checkTrieContents(oldDB, newDB *statedb.Database, root common.Hash) ([]common.Address, error) {
	oldTrie, err := statedb.NewSecureTrie(root, oldDB)
	if err != nil {
		return nil, err
	}

	newTrie, err := statedb.NewSecureTrie(root, newDB)
	if err != nil {
		return nil, err
	}

	diff, _ := statedb.NewDifferenceIterator(oldTrie.NodeIterator([]byte{}), newTrie.NodeIterator([]byte{}))
	iter := statedb.NewIterator(diff)

	var dirty []common.Address

	for iter.Next() {
		key := newTrie.GetKey(iter.Key)
		if key == nil {
			return nil, fmt.Errorf("no preimage found for hash %x", iter.Key)
		}

		dirty = append(dirty, common.BytesToAddress(key))
	}

	return dirty, nil
}

// CompareStateConsistency checks that all data of given two states.
func CompareStateConsistency(oldDB state.Database, newDB state.Database, root common.Hash) error {
	// Create and iterate a state trie rooted in a sub-node
	if _, err := oldDB.TrieDB().Node(root); err != nil {
		return err
	}
	oldState, err := state.New(root, oldDB)
	if err != nil {
		return err
	}

	if _, err := newDB.TrieDB().Node(root); err != nil {
		return err
	}
	newState, err := state.New(root, newDB)
	if err != nil {
		return err
	}

	oldIt := state.NewNodeIterator(oldState)
	newIt := state.NewNodeIterator(newState)
	for oldIt.Next() {
		if !newIt.Next() {
			return fmt.Errorf("newDB iterator finished early : oldIt.Hash(%v) oldIt.Parent(%v)", oldIt.Hash, oldIt.Parent)
		}

		if oldIt.Hash != newIt.Hash {
			return fmt.Errorf("mismatched hash oldIt.Hash : oldIt.Hash(%v) newIt.Hash(%v)", oldIt.Hash, newIt.Hash)
		}

		if oldIt.Parent != newIt.Parent {
			return fmt.Errorf("mismatched parent hash : oldIt.Parent(%v) newIt.Parent(%v)", oldIt.Parent, newIt.Parent)
		}
		logger.Info("Compare the state", "hash", oldIt.Hash, "parent", oldIt.Parent)
	}

	if newIt.Next() {
		return fmt.Errorf("oldDB iterator finished early  : newIt.Hash(%v) newIt.Parent(%v)", newIt.Hash, newIt.Parent)
	}

	return nil
}

func (bc *BlockChain) restartStateMigration() {
	if bc.db.InMigration() {
		number := bc.db.MigrationBlockNumber()

		block := bc.GetBlockByNumber(number)
		if block == nil {
			logger.Error("failed to get migration block number", "blockNumber", number)
			return
		}

		root := block.Root()
		logger.Warn("State migration is restarted", "blockNumber", number, "root", root.String())

		go bc.migrateState(root)
	}
}

func (bc *BlockChain) PrepareStateMigration() error {
	if bc.db.InMigration() || bc.prepareStateMigration {
		return errors.New("migration already started")
	}

	bc.prepareStateMigration = true
	currentBlock := bc.CurrentBlock().NumberU64()
	nextCommittedBlock := currentBlock + (DefaultBlockInterval - currentBlock%DefaultBlockInterval)
	logger.Warn("State migration is prepared", "migrationStartingBlockNumber", nextCommittedBlock)

	return nil
}

func (bc *BlockChain) checkStartStateMigration(number uint64, root common.Hash) {
	if bc.prepareStateMigration {
		logger.Info("State migration is started", "block", number, "root", root)

		if err := bc.StartStateMigration(number, root); err != nil {
			logger.Error("Failed to start state migration", "err", err)
		}

		bc.prepareStateMigration = false
	}
}

func (bc *BlockChain) StartStateMigration(number uint64, root common.Hash) error {
	// TODO-Klaytn Add internal status check routine
	if bc.db.InMigration() {
		return errors.New("migration already started")
	}

	if err := bc.db.CreateMigrationDBAndSetStatus(number); err != nil {
		return err
	}

	go bc.migrateState(root)

	return nil
}

func (bc *BlockChain) StopStateMigration() error {
	if !bc.db.InMigration() {
		return errors.New("not in migration")
	}

	bc.stopStateMigration <- struct{}{}

	return nil
}

// StatusStateMigration returns if it is in migration, the block number of in migration,
// number of committed blocks and number of pending blocks
func (bc *BlockChain) StatusStateMigration() (bool, uint64, int, int, float64) {
	return bc.db.InMigration(), bc.db.MigrationBlockNumber(), bc.committedCnt, bc.pendingCnt, bc.progress
}
