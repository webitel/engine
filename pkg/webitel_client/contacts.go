package webitel_client

import (
	"buf.build/gen/go/webitel/chat/protocolbuffers/go/messages"
	contacts_pb "buf.build/gen/go/webitel/webitel-go/protocolbuffers/go/contacts"
	"context"
)

type SearchContactsRequest = contacts_pb.SearchContactsRequest
type SearchContactsRequestNA = contacts_pb.SearchContactsNARequest
type LocateContactRequest = contacts_pb.LocateContactRequest
type InputContactRequest = contacts_pb.InputContactRequest
type Contact = contacts_pb.Contact
type ContactList = contacts_pb.ContactList
type LinkContactToConversation = messages.LinkContactToClientRequest

func (cli *Client) SearchContacts(token string, req *SearchContactsRequest) (*ContactList, error) {
	return cli.contactApi.SearchContacts(tokenContext(token), req)
}

func (cli *Client) SearchContactsNA(ctx context.Context, req *SearchContactsRequestNA) (*ContactList, error) {
	return cli.contactApi.SearchContactsNA(ctx, req)
}

func (cli *Client) LocateContact(token string, req *LocateContactRequest) (*Contact, error) {
	return cli.contactApi.LocateContact(tokenContext(token), req)
}

func (cli *Client) UpdateContact(token string, req *InputContactRequest) (*Contact, error) {
	return cli.contactApi.UpdateContact(tokenContext(token), req)
}

func (cli *Client) CreateContact(token string, req *InputContactRequest) (*Contact, error) {
	return cli.contactApi.CreateContact(tokenContext(token), req)
}

func (cli *Client) ContactLinkConversation(token string, req *LinkContactToConversation) error {
	_, err := cli.contactLinkApi.LinkContactToClient(tokenContext(token), req)

	return err
}
