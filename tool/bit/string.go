package bit

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func StringSpaceOnce(data string) string {
	res := make([]byte, 0)
	n := 0
	for i := 0; i < len(data); i++ {
		if data[i] != ' ' {
			res = append(res, data[i])
			n = 0
			continue
		} else if n == 0 {
			res = append(res, data[i])
			n = 1
		}
	}
	return string(res)
}

func Ltrim(in string) string {
	spaceCount := 0
	for i := 0; i < len(in); i++ {
		if in[i] == ' ' {
			spaceCount++
		} else {
			break
		}
	}
	out := fmt.Sprintf("%s", in[spaceCount:])
	return out
}

func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

func Strstr(src, dst string) bool {
	a := len(src)
	b := len(dst)
	if a <= 0 || b <= 0 || a < b {
		return false
	}
	i := 0
	for {
		if a-i < b {
			return false
		} else {
			if src[i:i+b] == dst[:] {
				return true
			}
		}
		i++
	}
}

// "+-*/"
//"+"
func StringAdd(s1, s2 string) string {
	i1, _ := strconv.ParseInt(s1, 10, 64)
	i2, _ := strconv.ParseInt(s2, 10, 64)
	i3 := i1 + i2
	return strconv.FormatInt(i3, 10)
}

//"-"
func StringSub(s1, s2 string) string {
	i1, _ := strconv.ParseInt(s1, 10, 64)
	i2, _ := strconv.ParseInt(s2, 10, 64)
	i3 := i1 - i2
	//log.Logger.Debug("DEBUG: ===== calu i1:%+v i2:%+v i3:%+v ", i1, i2, i3)
	return strconv.FormatInt(i3, 10)
}

// "*"
func StringMul(s1, s2 string) string {
	i1, _ := strconv.ParseInt(s1, 10, 64)
	i2, _ := strconv.ParseInt(s2, 10, 64)
	i3 := i1 * i2
	return strconv.FormatInt(i3, 10)
}

// "/"
func StringDiv(s1, s2 string) string {
	i1, _ := strconv.ParseInt(s1, 10, 64)
	i2, _ := strconv.ParseInt(s2, 10, 64)
	i3 := 100 * float64(i1) / float64(i2)
	return fmt.Sprintf("%0.2f", i3)
}

//100000000000000000=> 1    1 =>000000000000000001
func IntTo18BitDecimal(x string) string {
	lenght := len(x)
	head := ""
	for i := 0; i < 18-lenght; i++ {
		head = fmt.Sprintf("0%s", head)
	}
	y := fmt.Sprintf("%s%s", head, x)
	res := RTrimZero(y)
	return res
}

func StringFilAttofil(fil string, attoFil string) string {
	symbol := ""
	i1, _ := strconv.ParseInt(fil, 10, 64)
	i2, _ := strconv.ParseInt(attoFil, 10, 64)
	if i2 < 0 {
		//todo 如果i1小于0呢？？
		if i1 == 0 {
			i2 *= -1
			symbol = "-"
		} else {
			addAmount := math.Pow(10, 18)
			i2 += int64(addAmount)
			i1 -= 1
		}
	}
	//前边补0，后边去0，作为小数位
	s2 := RTrimZero(String18BitByLeftAddZero(strconv.FormatInt(i2, 10)))
	s := fmt.Sprintf("%v%v", symbol, i1)

	if len(s2) > 0 {
		s = fmt.Sprintf("%v.%v", s, s2)
	}
	return s
}

// 100000000000000000 => 1    155500000 => 1555
//去0
func RTrimZero(in string) string {
	zeroCount := 0
	for i := 0; i < len(in); i++ {
		if in[len(in)-i-1:len(in)-i] == "0" {
			zeroCount++
		} else {
			break
		}
	}
	out := fmt.Sprintf("%s", in[0:len(in)-zeroCount])
	return out
}

//前置位补0，至18位
func String18BitByLeftAddZero(in string) string {
	for i := len(in); i < 18; i++ {
		in = fmt.Sprintf("0%s", in)
	}
	return in
}

func String18BitByRightAddZero(in string) string {
	for i := len(in); i < 18; i++ {
		in = fmt.Sprintf("%s0", in)
	}
	return in
}

// "*"
func StringFloat64Mul(s1, s2 string) (string, error) {
	var err error
	var i1, i2 float64
	i1, err = strconv.ParseFloat(s1, 64)
	if err != nil {
		return "", err
	}
	i2, err = strconv.ParseFloat(s2, 64)
	if err != nil {
		return "", err
	}
	i3 := i1 * i2
	return fmt.Sprintf("%0.2f", i3), nil
}

// "/"
func StringFloat64Div(s1, s2 string) (string, error) {
	var err error
	var i1, i2 float64
	i1, err = strconv.ParseFloat(s1, 64)
	if err != nil {
		return "", err
	}
	i2, err = strconv.ParseFloat(s2, 64)
	if err != nil {
		return "", err
	}
	i3 := float64(i1) / float64(i2)
	return fmt.Sprintf("%0.2f", i3), nil
}

func TransFilToFIL(amount string) (mount string) {
	//log.Logger.Debug("******** WalletSettlementData transFilToFIL amount :%+v", amount)

	if len(amount) <= 18 {
		if amount == "0" {
			mount = "0.0"
		} else {
			mount = IntTo18BitDecimal(amount)
			mount = fmt.Sprintf("0.%s", mount)
			//log.Logger.Debug("******** WalletSettlementData transFilToFIL len<18 mount :%+v", mount)
		}
	} else {
		Fil := fmt.Sprintf("%s", amount[len(amount)-18:])
		Fil = RTrimZero(Fil)
		mount = fmt.Sprintf("%s.%s", amount[0:len(amount)-18], Fil)
		//log.Logger.Debug("******** WalletSettlementData transFilToFIL len>18 mount :%+v", mount)
	}

	return
}

func TransRewardToFilAndAttoFil(reward string) (fil, attoFil string) {
	if len(reward) <= 18 {
		return "0", reward
	} else {
		return reward[0 : len(reward)-18], reward[len(reward)-18:]
	}

}


func CalculateReward(beforeAmount string, nowAmount string) string {
	before := strings.Split(beforeAmount, ".")
	now := strings.Split(nowAmount, ".")
	before[1] = String18BitByRightAddZero(before[1])
	now[1] = String18BitByRightAddZero(now[1])
	fil := StringAdd(before[0], now[0])
	attoFil := StringAdd(before[1], now[1])
	//如果attoFil相加超过了18位
	if len(attoFil) > 18 {
		fil = StringAdd(fil, attoFil[0:1])
		attoFil = attoFil[1:]
	}
	if len(attoFil)<18{
		attoFil=String18BitByLeftAddZero(attoFil)
	}
	return fil + "." + attoFil
}