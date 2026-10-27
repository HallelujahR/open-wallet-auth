package client

import "testing"

func TestCallbackAllowListRequiresExactHTTPURL(t *testing.T) {
	client := Client{AllowedRedirectURIs: []string{"https://app.example.com/callback", "https://app.example.com/#/callback?source=auth", "http://localhost:5173/callback"}}
	for _, tt := range []struct {
		uri     string
		allowed bool
	}{
		{"", true}, {"https://app.example.com/callback", true}, {"https://app.example.com/#/callback?source=auth", true}, {"http://localhost:5173/callback", true},
		{"https://app.example.com.evil.test/callback", false}, {"https://app.example.com@evil.test/callback", false}, {"//evil.test/callback", false},
		{"javascript:alert(1)", false}, {"https://app.example.com/callback/extra", false}, {"https://app.example.com/callback?next=https://evil.test", false}, {"https://app.example.com/#/other", false},
	} {
		t.Run(tt.uri, func(t *testing.T) {
			if got := client.AllowsReturnURI(tt.uri); got != tt.allowed {
				t.Fatalf("allowed=%v, want %v", got, tt.allowed)
			}
		})
	}
	if (Client{}).AllowsReturnURI("https://app.example.com/callback") {
		t.Fatal("empty allow-list must reject redirects")
	}
}
