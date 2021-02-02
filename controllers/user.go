package controllers

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/models"
)

var userLog = logging.Logger("user-ctr-log")

type UserController struct {
	web.Controller
}

func (c *UserController) GetUserInfo() {
	userId := c.GetString("uid")

	if userId == "" {
		resp := models.UserInfoResp{
			Code: "faile",
			Msg:  "uid  is nil",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}

	userLog.Debug("DEBUG: GetUserInfo() user: %+v", userId)
	o := orm.NewOrm()
	userInfo := new(models.UserInfo)
	num, err := o.QueryTable("fly_user_info").Filter("user_id", userId).All(userInfo)
	//userLog.Debug("DEBUG: QueryRewardInfo() reward: %+v ", rewardInfo)
	if err != nil || num == 0 {
		resp := models.UserInfoResp{
			Code: "faile",
			Msg:  "haven't this user info",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		resp := models.UserInfoResp{
			Code: "ok",
			Msg:  "successful",
			Data: *userInfo,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
	}
}

func (c *UserController) GetUserDailyInfo() {
	userId := c.GetString("uid")
	index, err := c.GetInt("index")
	if err != nil {
		resp := models.UserDailyInfoResp{
			Code: "faile",
			Msg:  "get index  is faile",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	page, err := c.GetInt("page")
	if err != nil {
		resp := models.UserDailyInfoResp{
			Code: "faile",
			Msg:  "get page  is faile",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	if userId == "" {
		resp := models.UserDailyInfoResp{
			Code: "faile",
			Msg:  "uid  is nil",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}

	userLog.Debug("DEBUG: GetUserDailyInfo() user: %+v", userId)
	o := orm.NewOrm()
	userDailyInfo := make([]models.UserDailyRewardInfo, 0)
	num, err := o.QueryTable("fly_user_daily_reward_info").Filter("user_id", userId).All(&userDailyInfo)
	//userLog.Debug("DEBUG: QueryRewardInfo() reward: %+v ", rewardInfo)
	if err != nil {

	}
	if num == 0 {
		resp := models.UserDailyInfoResp{
			Code:       "ok",
			Msg:        "successful ",
			TotalCount: 0,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		result, pages := paging(userDailyInfo, page, index)

		resp := models.UserDailyInfoResp{
			Code:       "ok",
			Msg:        "successful",
			TotalCount: int(num),
			TotalPage:  int(pages),
			Data:       result,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
	}
}

func paging(infos []models.UserDailyRewardInfo, page, index int) ([]models.UserDailyRewardInfo, int) {
	pages := len(infos)/page + 1
	begin := (index - 1) * page
	end := (index) * page
	if index == pages {
		return infos[begin:], pages
	} else {
		return infos[begin:end], pages
	}

}
