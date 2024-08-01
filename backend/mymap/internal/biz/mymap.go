package biz

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
)

// 把不涉及数据库操作的业务逻辑封装在这里
type MyMapBiz struct {
	log *log.Helper
}

func NewMyMapBiz(logger log.Logger) *MyMapBiz {
	return &MyMapBiz{log: log.NewHelper(logger)}
}

// 获取驾驶信息
func (mmBiz *MyMapBiz) GetDriving(origin, destination string) (string, string, error) {
	key := "c4ef1c92a2cc184f95c5f74ad4255be5"
	api := "https://restapi.amap.com/v3/direction/driving"
	parameters := fmt.Sprintf("origin=%s&destination=%s&extensions=all&key=%s", origin, destination, key)
	url := api + "?" + parameters
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	// http响应body是数据流，要关闭
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body) //用io包读出来数据，body是[]byte的格式
	if err != nil {
		return "", "", err
	}

	// log.Info(string(body))
	// 把json格式的数据解析到结构体上
	ddResp := &DirectionDrivingResp{}
	if err := json.Unmarshal(body, ddResp); err != nil {
		return "", "", err
	}

	// 三.判断LSB的请求结果
	if ddResp.Infocode == "0" {
		return "", "", errors.New(ddResp.Info)
	}

	// 四.正确返回
	return ddResp.Route.Paths[0].Distance, ddResp.Route.Paths[0].Duration, nil
}

// 根据高德地图的返回格式定义的类型
type DirectionDrivingResp struct {
	Status   string `json:"status,omitempty"`
	Info     string `json:"info,omitempty"`
	Infocode string `json:"infocode,omitempty"`
	Count    string `json:"count,omitempty"`
	Route    struct {
		Origin      string `json:"origin,omitempty"`
		Destination string `json:"destination,omitempty"`
		Paths       []Path `json:"paths,omitempty"`
	} `json:"route"`
}

type Path struct {
	Distance string `json:"distance,omitempty"`
	Duration string `json:"duration,omitempty"`
	Strategy string `json:"strategy,omitempty"`
}
