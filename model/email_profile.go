package model

//TODO hide password

type EmailProfile struct {
	DomainRecord
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Schema      Lookup `json:"schema" db:"schema"`
	Enabled     bool   `json:"enabled" db:"enabled"`
	Host        string `json:"host" db:"host"`
	Login       string `json:"login" db:"login"`
	Password    string `json:"password" db:"password"`
	Mailbox     string `json:"mailbox" db:"mailbox"`
	SmtpPort    int    `json:"smtp_port" db:"smtp_port"`
	ImapPort    int    `json:"imap_port" db:"imap_port"`
}

type EmailProfilePatch struct {
	UpdatedBy Lookup
	UpdatedAt int64

	Name        *string `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Schema      *Lookup `json:"schema" db:"schema"`
	Enabled     *bool   `json:"enabled" db:"enabled"`
	Host        *string `json:"host" db:"host"`
	Login       *string `json:"login" db:"login"`
	Password    *string `json:"password" db:"-"`
	Mailbox     *string `json:"mailbox" db:"mailbox"`
	SmtpPort    *int    `json:"smtp_port" db:"smtp_port"`
	ImapPort    *int    `json:"imap_port" db:"imap_port"`
}

func (p *EmailProfile) Patch(patch *EmailProfilePatch) {
	p.UpdatedBy = &patch.UpdatedBy
	p.UpdatedAt = patch.UpdatedAt

	if patch.Name != nil {
		p.Name = *patch.Name
	}
	if patch.Description != nil {
		p.Description = *patch.Description
	}
	if patch.Schema != nil {
		p.Schema = *patch.Schema
	}
	if patch.Enabled != nil {
		p.Enabled = *patch.Enabled
	}
	if patch.Host != nil {
		p.Host = *patch.Host
	}
	if patch.Login != nil {
		p.Login = *patch.Login
	}
	if patch.Password != nil {
		p.Password = *patch.Password
	}
	if patch.Mailbox != nil {
		p.Mailbox = *patch.Mailbox
	}
	if patch.SmtpPort != nil {
		p.SmtpPort = *patch.SmtpPort
	}
	if patch.ImapPort != nil {
		p.ImapPort = *patch.ImapPort
	}
}

func (p EmailProfile) DefaultOrder() string {
	return "id"
}

func (p EmailProfile) AllowFields() []string {
	return []string{"id", "created_at", "created_by", "updated_at", "updated_by", "name", "enabled", "schema", "host",
		"mailbox", "description", "login", "smtp_port", "imap_port"}
}

func (p EmailProfile) DefaultFields() []string {
	return []string{"id", "name", "enabled", "schema", "host", "mailbox"}
}

func (p EmailProfile) EntityName() string {
	return "cc_email_profile_list"
}

func (p *EmailProfile) IsValid() *AppError {
	return nil //TODO
}

type SearchEmailProfile struct {
	ListRequest
}
