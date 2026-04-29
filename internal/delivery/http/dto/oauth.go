package dto

// OAuthStartResponse contains the provider authorization URL.
type OAuthStartResponse struct {
	Provider string `json:"provider"`
	AuthURL  string `json:"auth_url"`
	State    string `json:"state"`
}
