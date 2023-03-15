package auth_manager

import "context"

func (am *authManager) ProductLimit(ctx context.Context, token string, productName string) (int, error) {
	client, err := am.getAuthClient()
	if err != nil {
		return 0, err
	}

	return client.ProductLimit(ctx, token, productName)
}
