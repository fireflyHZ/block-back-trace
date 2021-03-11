package models

import "github.com/beego/beego/v2/client/orm"

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
