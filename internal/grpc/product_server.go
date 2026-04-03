package grpc

import (
	productpb "ims-database-util/internal/gen/productpb"
	"ims-database-util/internal/repository"
	"ims-database-util/internal/service"
)

type ProductServer struct {
	productpb.UnimplementedProductServiceServer
	service service.ProductService
}

func NewProductServer(s service.ProductService) *ProductServer {
	return &ProductServer{service: s}
}

func (s *ProductServer) StreamProducts(
	req *productpb.Empty,
	stream productpb.ProductService_StreamProductsServer,
) error {

	ctx := stream.Context()

	return s.service.StreamProducts(ctx, func(batch []repository.Product) error {
		for _, p := range batch {
			err := stream.Send(&productpb.Product{
				Id:          p.Id,
				Sku:         p.Sku,
				Name:        p.Name,
				Category:    p.Category,
				Price:       p.Price,
				Stock:       int32(p.Stock),
				Status:      p.Status,
				LastUpdated: p.LastUpdated.String(),
				UpdatedBy:   p.UpdatedBy,
				AddedBy:     p.AddedBy,
				UserId:      p.UserId,
				AddedAt:     p.AddedAt.String(),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}
