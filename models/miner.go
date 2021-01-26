package models

import "github.com/filecoin-project/specs-actors/actors/builtin/power"

type MinerPower struct {
	MinerPower  power.Claim
	TotalPower  power.Claim
	HasMinPower bool
}
