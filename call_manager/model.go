package call_manager

type SipRegistrationDeviceInfo struct {
	Uri                    string `json:"uri"`
	Id                     string `json:"id"`
	Type                   string `json:"type"`
	Sku                    string `json:"sku"`
	Status                 string `json:"status"`
	Name                   string `json:"name"`
	Serial                 string `json:"serial"`
	ComputerName           string `json:"computerName"`
	BoxBillingId           int    `json:"boxBillingId"`
	UseAsCommonPhone       bool   `json:"useAsCommonPhone"`
	LinePooling            string `json:"linePooling"`
	InCompanyNet           bool   `json:"inCompanyNet"`
	LastLocationReportTime string `json:"lastLocationReportTime"`
}

// SIPInfoResponse SIPInfo Response
type SIPInfoResponse struct {
	Username           string `json:"username"`
	Password           string `json:"password"`
	AuthorizationId    string `json:"authorizationId"`
	Domain             string `json:"domain"`
	OutboundProxy      string `json:"outboundProxy"`
	Transport          string `json:"transport"`
	Certificate        string `json:"certificate"`
	SwitchBackInterval int    `json:"switchBackInterval"`
}
