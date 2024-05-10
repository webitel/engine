package webitel_client

import (
	contacts_pb "buf.build/gen/go/webitel/webitel-go/protocolbuffers/go/contacts"
)

type SearchContactsRequest = contacts_pb.SearchContactsRequest
type LocateContactRequest = contacts_pb.LocateContactRequest
type InputContactRequest = contacts_pb.InputContactRequest
type Contact = contacts_pb.Contact
type ContactList = contacts_pb.ContactList

func (cli *Client) SearchContacts(token string, req *SearchContactsRequest) (*ContactList, error) {
	return cli.contactApi.SearchContacts(tokenContext(token), req)
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
