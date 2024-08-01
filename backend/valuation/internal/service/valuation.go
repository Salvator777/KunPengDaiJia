package service

import (
	"context"

	pb "valuation/api/valuation"
	"valuation/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
)

type ValuationService struct {
	pb.UnimplementedValuationServer
	Vbiz *biz.ValuationBiz
}

func NewValuationService(Vbiz *biz.ValuationBiz) *ValuationService {
	return &ValuationService{
		Vbiz: Vbiz,
	}
}

func (s *ValuationService) GetEstimatePrice(ctx context.Context, req *pb.GetEstimatePriceReq) (*pb.GetEstimatePriceReply, error) {
	distance, duration, err := s.Vbiz.GetDrivingInfoMyMap(req.Origin, req.Destination)
	if err != nil {
		return nil, err
	}
	price, err := s.Vbiz.GetPrice(context.Background(), distance, duration, 1, 10)
	if err != nil {
		return nil, errors.New(200, "PRICE ERROR", "cal price error")
	}
	return &pb.GetEstimatePriceReply{
		Origin:      req.Origin,
		Destination: req.Destination,
		Price:       price,
	}, nil
}
