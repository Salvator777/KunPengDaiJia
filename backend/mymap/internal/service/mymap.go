package service

import (
	"context"

	pb "mymap/api/mymap"
	"mymap/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
)

type MymapService struct {
	pb.UnimplementedMymapServer
	MMbiz *biz.MyMapBiz
}

func NewMymapService(MMbiz *biz.MyMapBiz) *MymapService {
	return &MymapService{MMbiz: MMbiz}
}

func (s *MymapService) GetDrivingInfo(ctx context.Context, req *pb.GetDrivingInfoReq) (*pb.GetDrivingInfoReply, error) {
	distance, duration, err := s.MMbiz.GetDriving(req.Origin, req.Destination)
	if err != nil {
		return nil, errors.New(200, "MymapService error", "map api error")
	}
	return &pb.GetDrivingInfoReply{
		Origin:      req.Origin,
		Destination: req.Destination,
		Distance:    distance,
		Duration:    duration,
	}, nil
}
