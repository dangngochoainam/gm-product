package api

import (
	"context"
	"log"

	productdtov1 "github.com/dangngochoainam/gm-core/application/product/v1/dto"
	productsvcv1 "github.com/dangngochoainam/gm-core/application/product/v1/service"
	"github.com/dangngochoainam/gm-product/configs"
	"github.com/dangngochoainam/gm-product/internal/usecase"
	"github.com/dangngochoainam/gopkg/copyhelper"
	"github.com/dangngochoainam/gopkg/loghelper"
	"google.golang.org/protobuf/types/known/emptypb"
)

type api struct {
	productsvcv1.UnimplementedProductServiceServer
	config         *configs.Config
	pbConverter    copyhelper.PbConverter
	productUsecase usecase.Product
}

func NewAPI(pbConverter copyhelper.PbConverter, config *configs.Config, productUsecase usecase.Product) productsvcv1.ProductServiceServer {
	return &api{
		config:         config,
		pbConverter:    pbConverter,
		productUsecase: productUsecase,
	}
}

func (s *api) SayHello(_ context.Context, in *productdtov1.HelloRequest) (*productdtov1.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &productdtov1.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *api) HealthCheck(ctx context.Context, in *emptypb.Empty) (*productdtov1.HelloReply, error) {
	loghelper.Logger.WithContext(ctx).Info("HealthCheck")
	return &productdtov1.HelloReply{Message: "OK"}, nil
}
