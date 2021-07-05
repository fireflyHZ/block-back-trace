package models

import (
	"github.com/beego/beego/v2/client/orm"
	"time"
)

func InsertPledgeMsg(p []*PreAndProveMessages) error {
	o := orm.NewOrm()
	for _, msg := range p {
		num, err := o.QueryTable("fly_pre_and_prove_messages").Filter("message_id", msg.MessageId).All(msg)
		if err != nil {
			return err
		}
		if num == 0 {
			_, err := o.Insert(msg)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (mbr *MineBlockRight) Insert() error {
	o := orm.NewOrm()
	num, err := o.QueryTable("fly_mine_block_right").Filter("miner_id", mbr.MinerId).Filter("epoch", mbr.Epoch).All(mbr)
	if err != nil {
		return err
	}
	if num == 0 {
		_, err := o.Insert(mbr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mbr *MineBlockRight) Update(t time.Time, value float64, winCount int64) error {
	o := orm.NewOrm()
	num, err := o.QueryTable("fly_mine_block_right").Filter("miner_id", mbr.MinerId).Filter("epoch", mbr.Epoch).All(mbr)
	if err != nil {
		return err
	}
	mbr.Missed = false
	mbr.Reward = value
	mbr.WinCount = winCount
	if num == 0 {
		mbr.Time = t
		mbr.UpdateTime = t
		_, err := o.Insert(mbr)
		if err != nil {
			return err
		}
	} else {
		_, err := o.Update(mbr)
		if err != nil {
			return err
		}
	}
	return nil
}
