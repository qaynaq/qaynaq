package coordinator

import (
	"context"
	"database/sql"
	"time"

	goshopify "github.com/bold-commerce/go-shopify/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

func (c *CoordinatorAPI) TestConnection(ctx context.Context, in *pb.TestConnectionRequest) (*pb.TestConnectionResponse, error) {
	driver := in.GetDriver()
	if driver != "postgres" && driver != "mysql" && driver != "sqlite" {
		return nil, status.Error(codes.InvalidArgument, "driver must be postgres, mysql, or sqlite")
	}

	connCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db, err := sql.Open(driver, in.GetConnectionString())
	if err != nil {
		return &pb.TestConnectionResponse{Ok: false, Error: err.Error()}, nil
	}
	defer db.Close()

	if err := db.PingContext(connCtx); err != nil {
		return &pb.TestConnectionResponse{Ok: false, Error: err.Error()}, nil
	}

	return &pb.TestConnectionResponse{Ok: true}, nil
}

func (c *CoordinatorAPI) TestShopifyConnection(ctx context.Context, in *pb.TestShopifyConnectionRequest) (*pb.TestShopifyConnectionResponse, error) {
	shopName := in.GetShopName()
	if shopName == "" {
		return nil, status.Error(codes.InvalidArgument, "shop_name is required")
	}

	apiAccessToken := in.GetApiAccessToken()
	if apiAccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "api_access_token is required")
	}

	connCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	app := goshopify.App{
		ApiKey:   "",
		Password: apiAccessToken,
	}
	client, err := goshopify.NewClient(app, shopName, "")
	if err != nil {
		return &pb.TestShopifyConnectionResponse{Ok: false, Error: err.Error()}, nil
	}

	shop, err := client.Shop.Get(connCtx, nil)
	if err != nil {
		return &pb.TestShopifyConnectionResponse{Ok: false, Error: err.Error()}, nil
	}

	return &pb.TestShopifyConnectionResponse{
		Ok:       true,
		ShopName: shop.Name,
	}, nil
}
