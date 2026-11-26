package sqlstore

import (
	"strings"
	"testing"

	"github.com/webitel/engine/model"
)

func TestAuditForm_TeamFirstSort(t *testing.T) {
	got, err := specFor(model.AuditForm{}).Expr("team_first")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !strings.Contains(got, ":TeamUserID::int8") {
		t.Errorf("not contains %s, got: %s", ":TeamUserID::int8", got)
	}

	if !strings.HasSuffix(got, "t.name ASC") {
		t.Errorf("not contains name ascending sort; got: %s", got)
	}
}
