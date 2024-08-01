package data

import (
	"valuation/internal/biz"
)

// 封装和PriceRule相关的数据操作的实现
type PriceRuleData struct {
	data *Data
}

func NewPriceRuleData(data *Data) biz.PriceRuleInterface {
	return &PriceRuleData{data: data}
}

// 用PriceRuleData 来实现 PriceRuleInterface 的接口
// 所以上面的new方法返回的是接口
// 根据城市和时间返回响应的规则
func (prd *PriceRuleData) GetRule(cityid uint, curr int) (*biz.PriceRule, error) {
	pr := &biz.PriceRule{}
	result := prd.data.Mdb.Where("city_id=? AND start_at <= ? AND ended_at > ?", cityid, curr, curr).First(pr)
	if result.Error != nil {
		return nil, result.Error
	}
	return pr, nil
}
