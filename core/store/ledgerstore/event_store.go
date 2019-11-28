// SPDX-License-Identifier: LGPL-3.0-or-later
// Copyright 2019 DNA Dev team
//
/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package ledgerstore

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/DNAProject/DNA/common"
	"github.com/DNAProject/DNA/common/log"
	"github.com/DNAProject/DNA/common/serialization"
	scom "github.com/DNAProject/DNA/core/store/common"
	"github.com/DNAProject/DNA/core/store/leveldbstore"
	"github.com/DNAProject/DNA/smartcontract/event"
)

//Saving event notifies gen by smart contract execution
type EventStore struct {
	dbDir string                     //Store path
	store *leveldbstore.LevelDBStore //Store handler
}

//NewEventStore return event store instance
func NewEventStore(dbDir string) (*EventStore, error) {
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	return &EventStore{
		dbDir: dbDir,
		store: store,
	}, nil
}

//NewBatch start event commit batch
func (this *EventStore) NewBatch() {
	this.store.NewBatch()
}

//SaveEventNotifyByTx persist event notify by transaction hash
func (this *EventStore) SaveEventNotifyByTx(txHash common.Uint256, notify *event.ExecuteNotify) error {
	result, err := json.Marshal(notify)
	if err != nil {
		return fmt.Errorf("json.Marshal error %s", err)
	}
	key := genEventNotifyByTxKey(txHash)
	this.store.BatchPut(key, result)
	return nil
}

//SaveEventNotifyByBlock persist transaction hash which have event notify to store
func (this *EventStore) SaveEventNotifyByBlock(height uint32, txHashs []common.Uint256) {
	key := genEventNotifyByBlockKey(height)
	values := common.NewZeroCopySink(nil)
	values.WriteUint32(uint32(len(txHashs)))
	for _, txHash := range txHashs {
		values.WriteHash(txHash)
	}
	this.store.BatchPut(key, values.Bytes())
}

//GetEventNotifyByTx return event notify by trasanction hash
func (this *EventStore) GetEventNotifyByTx(txHash common.Uint256) (*event.ExecuteNotify, error) {
	key := genEventNotifyByTxKey(txHash)
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	var notify event.ExecuteNotify
	if err = json.Unmarshal(data, &notify); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error %s", err)
	}
	return &notify, nil
}

//GetEventNotifyByBlock return all event notify of transaction in block
func (this *EventStore) GetEventNotifyByBlock(height uint32) ([]*event.ExecuteNotify, error) {
	key := genEventNotifyByBlockKey(height)
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewBuffer(data)
	size, err := serialization.ReadUint32(reader)
	if err != nil {
		return nil, fmt.Errorf("ReadUint32 error %s", err)
	}
	evtNotifies := make([]*event.ExecuteNotify, 0)
	for i := uint32(0); i < size; i++ {
		var txHash common.Uint256
		err = txHash.Deserialize(reader)
		if err != nil {
			return nil, fmt.Errorf("txHash.Deserialize error %s", err)
		}
		evtNotify, err := this.GetEventNotifyByTx(txHash)
		if err != nil {
			log.Errorf("getEventNotifyByTx Height:%d by txhash:%s error:%s", height, txHash.ToHexString(), err)
			continue
		}
		evtNotifies = append(evtNotifies, evtNotify)
	}
	return evtNotifies, nil
}

//CommitTo event store batch to store
func (this *EventStore) CommitTo() error {
	return this.store.BatchCommit()
}

//Close event store
func (this *EventStore) Close() error {
	return this.store.Close()
}

//ClearAll all data in event store
func (this *EventStore) ClearAll() error {
	this.NewBatch()
	iter := this.store.NewIterator(nil)
	for iter.Next() {
		this.store.BatchDelete(iter.Key())
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}
	return this.CommitTo()
}

//SaveCurrentBlock persist current block height and block hash to event store
func (this *EventStore) SaveCurrentBlock(height uint32, blockHash common.Uint256) {
	key := this.getCurrentBlockKey()
	value := common.NewZeroCopySink(nil)
	value.WriteHash(blockHash)
	value.WriteUint32(height)
	this.store.BatchPut(key, value.Bytes())
}

//GetCurrentBlock return current block hash, and block height
func (this *EventStore) GetCurrentBlock() (common.Uint256, uint32, error) {
	key := this.getCurrentBlockKey()
	data, err := this.store.Get(key)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	reader := bytes.NewReader(data)
	blockHash := common.Uint256{}
	err = blockHash.Deserialize(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	height, err := serialization.ReadUint32(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	return blockHash, height, nil
}

func (this *EventStore) getCurrentBlockKey() []byte {
	return []byte{byte(scom.SYS_CURRENT_BLOCK)}
}

func genEventNotifyByBlockKey(height uint32) []byte {
	key := make([]byte, 5, 5)
	key[0] = byte(scom.EVENT_NOTIFY)
	binary.LittleEndian.PutUint32(key[1:], height)
	return key
}

func genEventNotifyByTxKey(txHash common.Uint256) []byte {
	data := txHash.ToArray()
	key := make([]byte, 1+len(data))
	key[0] = byte(scom.EVENT_NOTIFY)
	copy(key[1:], data)
	return key
}
