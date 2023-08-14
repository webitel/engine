package webitel_client

import (
	"context"
	"errors"
	auth_pb "github.com/webitel/engine/pkg/webitel_client/api"
	"google.golang.org/grpc/metadata"
)

func (cli *Client) ProductLimit(ctx context.Context, token string, productName string) (int, error) {
	header := metadata.New(map[string]string{"x-webitel-access": token})
	outCtx := metadata.NewOutgoingContext(ctx, header)
	tenant, err := cli.customerApi.GetCustomer(outCtx, &auth_pb.GetCustomerRequest{})

	if err != nil {
		return 0, err
	}

	if tenant.Customer == nil {
		return 0, errors.New("customer is empty")
	}

	var limitMax int32

	for _, grant := range tenant.Customer.GetLicense() {
		if grant.Product != productName {
			continue // Lookup productName only !
		}
		if errs := grant.GetStatus().GetErrors(); len(errs) != 0 {
			// Also, ignore single 'product exhausted' (remain < 1) error
			// as we do not consider product user assignments here ...
			if !(len(errs) == 1 && errs[0] == "product exhausted") {
				continue // Currently invalid
			}
		}
		if limitMax < grant.Remain {
			limitMax = grant.Remain
		}
	}

	if limitMax == 0 {
		// FIXME: No CHAT product(s) issued !
		return 0, errors.New("")
	}

	return int(limitMax), nil
}
