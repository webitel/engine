package logger

import (
	"context"
	"github.com/webitel/engine/model"
	"testing"
)

type sender struct {
}
type session struct {
}

func (s *session) GetUserIp() string {
	return "test"
}

func (s *session) GetUserId() int64 {
	return 10
}
func (s *session) GetDomainId() int64 {
	return 1
}

var (
	o = struct {
		A int
		B string
		C struct {
			D float32
		}
	}{}
)

func (s *sender) Send(ctx context.Context, exchange string, rk string, body []byte) error {
	// TODO
	return nil
}

func TestLogger(t *testing.T) {
	logger, err := New(&sender{})
	if err != nil {
		t.Fatal(err.Error())
	}

	testLogger(logger, t)
}

func BenchmarkLogger(t *testing.B) {
	logger, err := New(&sender{})
	if err != nil {
		t.Fatal(err.Error())
	}
	for i := 0; i < t.N; i++ {
		testLogger(logger, t)
	}
}

func testLogger(logger *Audit, t testing.TB) {
	ctx := context.TODO()

	err := logger.Update(ctx, &session{}, model.PERMISSION_SCOPE_SCHEMA, 1, &o)
	if err != nil {
		t.Fatal(err.Error())
	}

}
