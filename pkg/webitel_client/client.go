package webitel_client

import (
	"context"
	"fmt"
	auth_pb "github.com/webitel/engine/pkg/webitel_client/api"
	contacts_pb "github.com/webitel/engine/pkg/webitel_client/api/contacts"
	"github.com/webitel/engine/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	_ "github.com/mbobakov/grpc-consul-resolver"
)

var (
	webitelServiceName = "go.webitel.app"
)

type Client struct {
	session     utils.ObjectCache
	contactApi  contacts_pb.ContactsClient
	authApi     auth_pb.AuthClient
	customerApi auth_pb.CustomersClient
	conn        *grpc.ClientConn
}

func New(cacheSize int, cacheTime int64, consulTarget string) (*Client, error) {
	conn, err := grpc.Dial(fmt.Sprintf("consul://%s/%s?wait=14s", consulTarget, webitelServiceName),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:        conn,
		session:     utils.NewLruWithParams(cacheSize, "sessions", cacheTime, ""), //TODO session from config ?
		contactApi:  contacts_pb.NewContactsClient(conn),
		authApi:     auth_pb.NewAuthClient(conn),
		customerApi: auth_pb.NewCustomersClient(conn),
	}, nil

}

func (cli *Client) Stop() {
	cli.conn.Close()
}

func (cli *Client) Test() {
	header := metadata.New(map[string]string{"x-webitel-access": "SUPER"})
	outCtx := metadata.NewOutgoingContext(context.TODO(), header)

	res, err := cli.contactApi.SearchContacts(outCtx, &contacts_pb.SearchContactsRequest{
		Page:   0,
		Size:   0,
		Q:      "",
		Sort:   nil,
		Fields: nil,
		Id:     nil,
		Qin:    nil,
		Mode:   0,
	})

	if err != nil {
		panic(err.Error())
	}

	fmt.Println(res)
}
