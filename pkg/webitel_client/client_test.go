package webitel_client

import (
	"context"
	"testing"
)

func TestContactManager(t *testing.T) {
	t.Log("TestContactManager")

	am, err := New(35000, 60, "10.9.8.111:8500")
	if err != nil {
		t.Fatal(err.Error())
	}

	token := "SUPER"

	sess, err := am.GetSession(token)
	if err != nil {
		t.Fatal(err.Error())
	}

	if sess == nil {
		t.Fatal("empty session")
	}

	count, err := am.ProductLimit(context.Background(), token, LicenseCallCenter)

	t.Logf("count license %d", count)

	list, err := am.SearchContacts(token, &SearchContactsRequest{
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
		t.Fatal(err.Error())
	}

	t.Log(list)

	defer am.Stop()
}
