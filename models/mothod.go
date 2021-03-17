package models

import (
	"github.com/beego/beego/v2/client/orm"
	"time"
)

func (p *PreAndProveMessages) Insert() error {
	o := orm.NewOrm()
	num, err := o.QueryTable("fly_pre_and_prove_messages").Filter("message_id", p.MessageId).All(p)
	if err != nil {
		return err
	}
	if num == 0 {
		_, err := o.Insert(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mb *NetMinerAndBlock) Insert() error {
	o := orm.NewOrm()
	num, err := o.QueryTable("fly_net_miner_and_block").Filter("miner_id", mb.MinerId).Filter("epoch", mb.Epoch).All(mb)
	if err != nil {
		return err
	}
	if num == 0 {
		_, err := o.Insert(mb)
		if err != nil {
			return err
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
