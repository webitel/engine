package model

type Region struct {
	Id          int64   `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Timezone    Lookup  `json:"timezone" db:"timezone"`
}

type SearchRegion struct {
	ListRequest
	Ids         []int64
	TimezoneIds []uint32
	Name        *string
	Description *string
}

type RegionPatch struct {
	Name        *string
	Description *string
	Timezone    *Lookup
}

func (r Region) AllowFields() []string {
	return r.DefaultFields()
}

func (r Region) DefaultOrder() string {
	return "+name"
}

func (r Region) DefaultFields() []string {
	return []string{"id", "name", "description", "timezone"}
}

func (r Region) EntityName() string {
	return "region_list"
}

func (r *Region) Patch(patch *RegionPatch) {

	if patch.Name != nil {
		r.Name = *patch.Name
	}

	if patch.Description != nil {
		r.Description = patch.Description
	}

	if patch.Timezone != nil {
		r.Timezone = *patch.Timezone
	}
}

// Todo
func (r *Region) IsValid() *AppError {
	return nil
}
