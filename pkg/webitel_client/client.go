package webitel_client

import (
	chgrpc "buf.build/gen/go/webitel/chat/grpc/go/messages/messagesgrpc"
	gogrpc "buf.build/gen/go/webitel/webitel-go/grpc/go/_gogrpc"
	congrpc "buf.build/gen/go/webitel/webitel-go/grpc/go/contacts/contactsgrpc"
	conproto "buf.build/gen/go/webitel/webitel-go/protocolbuffers/go/contacts"
	"context"
	"fmt"

	"github.com/webitel/engine/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	_ "github.com/mbobakov/grpc-consul-resolver"
)

var (
	webitelServiceName = "go.webitel.app"
)

type Client struct {
	session        utils.ObjectCache
	contactApi     congrpc.ContactsClient
	contactLinkApi chgrpc.ContactLinkingServiceClient
	authApi        gogrpc.AuthClient
	customerApi    gogrpc.CustomersClient
	conn           *grpc.ClientConn
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
		conn:           conn,
		session:        utils.NewLruWithParams(cacheSize, "sessions", cacheTime, ""), //TODO session from config ?
		contactApi:     congrpc.NewContactsClient(conn),
		contactLinkApi: chgrpc.NewContactLinkingServiceClient(conn),
		authApi:        gogrpc.NewAuthClient(conn),
		customerApi:    gogrpc.NewCustomersClient(conn),
	}, nil

}

func (cli *Client) Stop() {
	cli.conn.Close()
}

func (cli *Client) Test() {
	header := metadata.New(map[string]string{"x-webitel-access": "SUPER"})
	outCtx := metadata.NewOutgoingContext(context.TODO(), header)

	res, err := cli.contactApi.SearchContacts(outCtx, &conproto.SearchContactsRequest{
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
