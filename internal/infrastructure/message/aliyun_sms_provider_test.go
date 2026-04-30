package message

import (
	"net/url"
	"strings"
	"testing"
)

func TestCanonicalQueryExcludesSignatureAndSorts(t *testing.T) {
	values := url.Values{}
	values.Set("Signature", "ignored")
	values.Set("Version", "2017-05-25")
	values.Set("Action", "SendSms")

	got := canonicalQuery(values)
	if got != "Action=SendSms&Version=2017-05-25" {
		t.Fatalf("unexpected canonical query: %s", got)
	}
}

func TestPercentEncodeUsesAliyunRules(t *testing.T) {
	got := percentEncode("a b~c*d")
	if got != "a%20b~c%2Ad" {
		t.Fatalf("unexpected encoded value: %s", got)
	}
}

func TestAliyunSignatureReturnsBase64(t *testing.T) {
	values := url.Values{}
	values.Set("Action", "SendSms")
	values.Set("AccessKeyId", "test")

	got := aliyunSignature(values, "secret")
	if got == "" || strings.Contains(got, " ") {
		t.Fatalf("unexpected signature: %q", got)
	}
}
