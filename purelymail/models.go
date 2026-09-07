package purelymail

type Empty struct{}

type UserInfo struct {
	EnableSearchIndexing           bool          `json:"enableSearchIndexing"`
	RecoveryEnabled                bool          `json:"recoveryEnabled"`
	RequireTwoFactorAuthentication bool          `json:"requireTwoFactorAuthentication"`
	EnableSpamFiltering            bool          `json:"enableSpamFiltering"`
	ResetMethods                   []ResetMethod `json:"resetMethods"`
}

type ResetMethod struct {
	Type          string `json:"type"`
	Target        string `json:"target"`
	Description   string `json:"description"`
	AllowMFAReset bool   `json:"allowMfaReset"`
}

type RoutingRule struct {
	ID              int64    `json:"id"`
	DomainName      string   `json:"domainName"`
	Prefix          bool     `json:"prefix"`
	MatchUser       string   `json:"matchUser"`
	TargetAddresses []string `json:"targetAddresses"`
	Catchall        bool     `json:"catchall"`
}

type Domain struct {
	Name                  string     `json:"name"`
	AllowAccountReset     bool       `json:"allowAccountReset"`
	SymbolicSubaddressing bool       `json:"symbolicSubaddressing"`
	IsShared              bool       `json:"isShared"`
	DNSSummary            DNSSummary `json:"dnsSummary"`
}

type DNSSummary struct {
	PassesMX    bool `json:"passesMx"`
	PassesSPF   bool `json:"passesSpf"`
	PassesDKIM  bool `json:"passesDkim"`
	PassesDMARC bool `json:"passesDmarc"`
}

type listUsersResponse struct {
	Users []string `json:"users"`
}
type listDomainsResponse struct {
	Domains []Domain `json:"domains"`
}
type listRulesResponse struct {
	Rules []RoutingRule `json:"rules"`
}
type listResetResponse struct {
	Users []ResetMethod `json:"users"`
}
type ownershipResponse struct {
	Code string `json:"code"`
}
type appPasswordResponse struct {
	AppPassword string `json:"appPassword"`
}
type creditResponse struct {
	Credit string `json:"credit"`
}
