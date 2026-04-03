package grpc

import (
	customerpb "ims-database-util/internal/gen/customerpb"
	"ims-database-util/internal/repository"
	"ims-database-util/internal/service"
)

type CustomerServer struct {
	customerpb.UnimplementedCustomerServiceServer
	service service.CustomerService
}

func NewCustomerServer(s service.CustomerService) *CustomerServer {
	return &CustomerServer{service: s}
}

func (s *CustomerServer) StreamCustomers(
	req *customerpb.Empty,
	stream customerpb.CustomerService_StreamCustomersServer,
) error {
	ctx := stream.Context()

	return s.service.StreamCustomers(ctx, func(batch []repository.Customer) error {
		for _, c := range batch {
			err := stream.Send(&customerpb.Customer{
				Id:        c.Id,
				FirstName: c.FirstName,
				LastName:  c.LastName,
				Email:     c.Email,
				Phone:     c.Phone,
				Address:   c.Address,
				Status:    c.Status,
				UserId:    c.UserId,
				CreatedAt: c.CreatedAt.String(),
				UpdatedAt: c.UpdatedAt.String(),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}
