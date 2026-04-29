package main

import "testing"

func TestBearerToken(t *testing.T) {
	token, ok := bearerToken("Bearer access-token")
	if !ok {
		t.Fatal("expected bearer token")
	}
	if token != "access-token" {
		t.Fatalf("unexpected token: %s", token)
	}
}

func TestBearerTokenRejectsMalformedHeader(t *testing.T) {
	if _, ok := bearerToken("Basic access-token"); ok {
		t.Fatal("expected malformed header to be rejected")
	}
	if _, ok := bearerToken("Bearer "); ok {
		t.Fatal("expected empty bearer token to be rejected")
	}
}
