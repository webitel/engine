package model

import (
	"encoding/json"
	"golang.org/x/oauth2"
	"strings"
)

const (
	MailGmail   = "gmail"
	MailOutlook = "outlook"
)

const (
	EmailAuthTypeOAuth2 = "oauth2"
	EmailAuthTypePlain  = "plain"
)

type OAuth2Config struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

type MailProfileParams struct {
	OAuth2 *OAuth2Config `json:"oauth2"`
}

type EmailProfile struct {
	DomainRecord
	Name          string             `json:"name" db:"name"`
	Description   string             `json:"description" db:"description"`
	Schema        Lookup             `json:"schema" db:"schema"`
	Enabled       bool               `json:"enabled" db:"enabled"`
	Login         string             `json:"login" db:"login"`
	Password      string             `json:"password" db:"password"`
	Mailbox       string             `json:"mailbox" db:"mailbox"`
	SmtpHost      string             `json:"smtp_host" db:"smtp_host"`
	SmtpPort      int                `json:"smtp_port" db:"smtp_port"`
	ImapHost      string             `json:"imap_host" db:"imap_host"`
	ImapPort      int                `json:"imap_port" db:"imap_port"`
	FetchInterval int32              `json:"fetch_interval" db:"fetch_interval"`
	FetchError    *string            `json:"fetch_error" db:"fetch_error"`
	State         string             `json:"state" db:"state"`
	ActivityAt    int64              `json:"activity_at" db:"activity_at"`
	Params        *MailProfileParams `json:"params" db:"params"`
	AuthType      string             `json:"auth_type" db:"auth_type"`
	Listen        bool               `json:"listen" db:"listen"`
	Logged        bool               `json:"logged" db:"logged"`
}

func (p *EmailProfile) oauthConfig() *OAuth2Config {
	if p.Params != nil && p.Params.OAuth2 != nil {
		return p.Params.OAuth2
	}

	return nil
}

func (p *EmailProfile) Oauth() (oauth2.Config, AppError) {
	config := p.oauthConfig()
	if config == nil {
		// TODO
		return oauth2.Config{}, nil
	}

	if strings.Index(p.ImapHost, MailGmail+".com") > -1 {

	} else if strings.Index(p.ImapHost, MailOutlook) == 0 {
		return oauth2.Config{
			ClientID:     config.ClientId,
			ClientSecret: config.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.microsoftonline.com/organizations/oauth2/v2.0/authorize",
				TokenURL: "https://login.microsoftonline.com/organizations/oauth2/v2.0/token",
			},
			RedirectURL: config.RedirectURL, //"https://dev.webitel.com/endpoint/oauth2/outlook/callback",
			Scopes: []string{
				"https://outlook.office.com/User.Read",
				"https://outlook.office.com/IMAP.AccessAsUser.All",
				"https://outlook.office.com/SMTP.Send",
				"offline_access",
			},
		}, nil
	}

	return oauth2.Config{}, NewBadRequestError("email.profile.valid.oauth", "Not found oauth config")
}

type EmailProfileLogin struct {
	AuthType    string `json:"auth_type" db:"auth_type"`
	RedirectUrl string `json:"redirect_url" db:"redirect_url"`
	Cookie      map[string]string
}

type EmailProfilePatch struct {
	UpdatedBy Lookup
	UpdatedAt int64

	Name          *string `json:"name" db:"name"`
	Description   *string `json:"description" db:"description"`
	Schema        *Lookup `json:"schema" db:"schema"`
	Enabled       *bool   `json:"enabled" db:"enabled"`
	Login         *string `json:"login" db:"login"`
	Password      *string `json:"password" db:"-"`
	Mailbox       *string `json:"mailbox" db:"mailbox"`
	SmtpHost      *string `json:"smtp_host" db:"smtp_host"`
	SmtpPort      *int    `json:"smtp_port" db:"smtp_port"`
	ImapHost      *string `json:"imap_host" db:"imap_host"`
	ImapPort      *int    `json:"imap_port" db:"imap_port"`
	FetchInterval *int32  `json:"fetch_interval" db:"fetch_interval"`
	Listen        *bool   `json:"listen" db:"listen"`
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
	if patch.SmtpHost != nil {
		p.SmtpHost = *patch.SmtpHost
	}
	if patch.ImapHost != nil {
		p.ImapHost = *patch.ImapHost
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

	if patch.FetchInterval != nil {
		p.FetchInterval = *patch.FetchInterval
	}

	if patch.Listen != nil {
		p.Listen = *patch.Listen
	}
}

func (p EmailProfile) DefaultOrder() string {
	return "id"
}

func (p EmailProfile) AllowFields() []string {
	return []string{"id", "created_at", "created_by", "updated_at", "updated_by", "name", "enabled", "schema", "smtp_host",
		"mailbox", "description", "login", "smtp_port", "imap_port", "password", "imap_host", "fetch_interval", "fetch_error",
		"state", "activity_at", "listen", "logged"}
}

func (p EmailProfile) DefaultFields() []string {
	return []string{"id", "name", "enabled", "schema", "mailbox", "state", "fetch_error", "listen"}
}

func (p EmailProfile) EntityName() string {
	return "cc_email_profile_list"
}

func (p *EmailProfile) IsValid() AppError {
	return nil //TODO
}

type SearchEmailProfile struct {
	ListRequest
}

func (p *MailProfileParams) Json() []byte {
	if p == nil {
		return []byte("{}")
	}

	data, _ := json.Marshal(p)
	return data
}
