package model

type SchemeVersion struct {
	Id        int64   `json:"id,omitempty" db:"id"`
	SchemeId  int64   `json:"schemaId,omitempty" db:"scheme_id"`
	CreatedAt int64   `json:"createdAt" db:"created_at"`
	CreatedBy Lookup  `json:"createdBy" db:"created_by"`
	Scheme    []byte  `json:"scheme,omitempty" db:"scheme"`
	Payload   []byte  `json:"payload,omitempty" db:"payload"`
	Version   int64   `json:"version,omitempty" db:"version"`
	Note      *string `json:"note,omitempty" db:"note"`
}

type SearchSchemeVersion struct {
	SchemeId int64
	ListRequest
}

var SchemeVersionFields = struct {
	Id        string
	SchemeId  string
	CreatedAt string
	CreatedBy string
	Scheme    string
	Payload   string
	Version   string
	Note      string
}{
	Id:        "id",
	SchemeId:  "scheme_id",
	CreatedAt: "created_at",
	CreatedBy: "created_by",
	Scheme:    "scheme",
	Payload:   "payload",
	Version:   "version",
	Note:      "note",
}
