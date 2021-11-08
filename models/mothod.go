package models

import (
	"github.com/beego/beego/v2/client/orm"
	"time"
)

func InsertPledgeMsg(p []*PreAndProveMessages) error {
	o := orm.NewOrm()
	for _, msg := range p {
		num, err := o.QueryTable("fly_pre_and_prove_messages").Filter("message_id", msg.MessageId).Filter("sector_number", msg.SectorNumber).All(msg)
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

func (mbr *MineBlockRight) Insert() (bool, error) {
	o := orm.NewOrm()
	num, err := o.QueryTable("fly_mine_block_right").Filter("miner_id", mbr.MinerId).Filter("epoch", mbr.Epoch).All(mbr)
	if err != nil {
		return true, err
	}
	if num == 0 {
		_, err = o.Insert(mbr)
		if err != nil {
			return true, err
		}
		return true, nil
	} else {
		return false, nil
	}

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

func (msg *ExpendMessages) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(msg)
	return err
}

func UpdateWalletsInfo(newWalletInfos map[string]*WalletInfo) error {
	o := orm.NewOrm()
	wallets := make([]WalletInfo, 0)

	for id, info := range newWalletInfos {
		num, err := o.QueryTable("fly_wallet_info").Filter("wallet_id", id).OrderBy("-create_time").All(&wallets)
		if err != nil {
			return err
		}

		if num == 0 {
			_, err = o.Insert(info)
			if err != nil {
				return err
			}
		} else {
			if info.Balance != wallets[0].Balance {
				_, err = o.Insert(info)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (mr *MineRight) Insert() error {
	o := orm.NewOrm()
	num, err := o.QueryTable("fly_mine_right").Filter("miner_id", mr.MinerId).Filter("wallet", mr.Wallet).Filter("epoch", mr.Epoch).All(mr)
	if err != nil {
		return err
	}
	if num == 0 {
		_, err = o.Insert(mr)
		if err != nil {
			return err
		}
		return nil
	} else {
		return nil
	}

}
