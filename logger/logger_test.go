package logger

import (
	"context"
	"testing"
)

type sender struct {
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

func (s *sender) Send(ctx context.Context, domainId int64, object string, body []byte) error {
	// TODO
	return nil
}

func TestLogger(t *testing.T) {
	logger, err := New("10.9.8.111:8500", &sender{})
	if err != nil {
		t.Fatal(err.Error())
	}

	testLogger(logger, t)
}

func BenchmarkLogger(t *testing.B) {
	logger, err := New("10.9.8.111:8500", &sender{})
	if err != nil {
		t.Fatal(err.Error())
	}
	for i := 0; i < t.N; i++ {
		testLogger(logger, t)
	}
}

func testLogger(logger *Api, t testing.TB) {
	ctx := context.TODO()

	err := logger.Audit(ctx, 1, "TODO", &o)
	if err != nil {
		t.Fatal(err.Error())
	}

}
