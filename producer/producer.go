// Copyright 2017~2022 The Bottos Authors
// This file is part of the Bottos Chain library.
// Created by Rocket Core Team of Bottos.

//This program is free software: you can distribute it and/or modify
//it under the terms of the GNU General Public License as published by
//the Free Software Foundation, either version 3 of the License, or
//(at your option) any later version.

//This program is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//GNU General Public License for more details.

//You should have received a copy of the GNU General Public License
// along with bottos.  If not, see <http://www.gnu.org/licenses/>.

/*
 * file description:  producer entry
 * @Author:
 * @Date:   2017-12-06
 * @Last Modified by:
 * @Last Modified time:
 */
package producer

import (
	"fmt"
	"time"

	"github.com/bottos-project/core/chain"
	"github.com/bottos-project/core/common"
	"github.com/bottos-project/core/common/types"
	"github.com/bottos-project/core/config"
	"github.com/bottos-project/core/db"
	"github.com/bottos-project/core/role"
)

type Reporter struct {
	isReporting bool
	core        chain.BlockChainInterface
	db          *db.DBService
}
type ReporterRepo interface {
	Woker(Trxs []*types.Transaction) *types.Block
	VerifyTrxs(Trxs []*types.Transaction) error
	IsReady() bool
}

func (p *Reporter) GetSlotAtTime(current time.Time) uint32 {
	firstSlotTime := p.GetSlotTime(1)

	if common.NowToSeconds(current) < firstSlotTime {
		return 0
	}
	return uint32(common.NowToSeconds(current)-firstSlotTime)/config.DEFAULT_BLOCK_INTERVAL + 1
}

func (p *Reporter) GetSlotTime(slotNum uint32) uint64 {

	if slotNum == 0 {
		return 0
	}
	interval := config.DEFAULT_BLOCK_INTERVAL

	object, err := role.GetChainStateObjectRole(p.db)
	if err != nil {
		return 0
	}
	genesisTime := p.core.GenesisTimestamp()
	if object.LastBlockNum == 0 {

		return genesisTime + uint64(slotNum*interval)
	}
	headBlockAbsSlot := common.GetSecondSincEpoch(object.LastBlockTime, genesisTime) / uint64(interval)
	headSlotTime := headBlockAbsSlot * uint64(interval)
	return headSlotTime + uint64(slotNum*interval)
}
func New(b chain.BlockChainInterface, db *db.DBService) ReporterRepo {
	return &Reporter{core: b, db: db}
}
func (p *Reporter) isEligible() bool {
	return true
}
func (p *Reporter) isReady() bool {
	if p.isReporting == true {
		return false
	}
	return true
	slotTime := p.GetSlotTime(1)
	fmt.Println(slotTime)
	if slotTime >= common.NowToSeconds(time.Now()) {
		return true
	}
	return false
}
func (p *Reporter) isMyTurn() bool {
	return true
}
func (p *Reporter) IsReady() bool {
	if p.isEligible() && p.isReady() && p.isMyTurn() {
		p.isReporting = true
		return true
	}
	return false
}
func (p *Reporter) Woker(trxs []*types.Transaction) *types.Block {

	now := time.Now()
	slot := p.GetSlotAtTime(now)
	scheduledTime := p.GetSlotTime(slot)
	fmt.Println("Woker", scheduledTime, slot)
	accountName, err1 := role.GetScheduleDelegateRole(p.db, slot)
	if err1 != nil {
		return nil // errors.New("report Block failed")
	}

	block, err := p.reportBlock(accountName, trxs)
	if err != nil {
		return nil // errors.New("report Block failed")
	}

	fmt.Println("brocasting block", block)
	return block
}

func (p *Reporter) StartTag() error {
	//p.core.

	return nil

}
func (p *Reporter) VerifyTrxs(trxs []*types.Transaction) error {

	return nil
}

//func reportBlock(reportTime time.Time, reportor role.Delegate) *types.Block {
func (p *Reporter) reportBlock(accountName string, trxs []*types.Transaction) (*types.Block, error) {
	head := types.NewHeader()
	head.PrevBlockHash = p.core.HeadBlockHash().Bytes()
	head.Number = p.core.HeadBlockNum() + 1
	head.Timestamp = p.core.HeadBlockTime() + uint64(config.DEFAULT_BLOCK_INTERVAL)
	head.Delegate = []byte("my")
	block := types.NewBlock(head, trxs)
	block.Header.DelegateSign = block.Sign("123").Bytes()
	p.isReporting = false
	return block, nil

}
