package message

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	phoneusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/phone"
)

// AliyunSMSConfig configures the Aliyun SendSms RPC API.
// AliyunSMSConfig 配置阿里云 SendSms RPC API。
type AliyunSMSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
	RegionID        string
	Endpoint        string
}

// AliyunSMSProvider sends verification SMS messages through Aliyun.
// AliyunSMSProvider 通过阿里云短信服务发送验证码。
type AliyunSMSProvider struct {
	cfg        AliyunSMSConfig
	httpClient *http.Client
}

// NewAliyunSMSProvider creates an Aliyun SMS provider.
// NewAliyunSMSProvider 创建阿里云短信发送适配器。
func NewAliyunSMSProvider(cfg AliyunSMSConfig) *AliyunSMSProvider {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://dysmsapi.aliyuncs.com"
	}
	if cfg.RegionID == "" {
		cfg.RegionID = "cn-hangzhou"
	}
	return &AliyunSMSProvider{cfg: cfg, httpClient: &http.Client{Timeout: 10 * time.Second}}
}

// SendSMS sends a template SMS with the code as TemplateParam.code.
// SendSMS 使用模板短信发送验证码，模板变量为 code。
func (p *AliyunSMSProvider) SendSMS(ctx context.Context, msg phoneusecase.SMSMessage) error {
	if err := p.validate(); err != nil {
		return err
	}
	templateParam, err := json.Marshal(map[string]string{"code": msg.Code})
	if err != nil {
		return err
	}
	values := url.Values{}
	values.Set("Action", "SendSms")
	values.Set("Version", "2017-05-25")
	values.Set("RegionId", p.cfg.RegionID)
	values.Set("PhoneNumbers", msg.Phone)
	values.Set("SignName", p.cfg.SignName)
	values.Set("TemplateCode", p.cfg.TemplateCode)
	values.Set("TemplateParam", string(templateParam))
	values.Set("Format", "JSON")
	values.Set("SignatureMethod", "HMAC-SHA1")
	values.Set("SignatureVersion", "1.0")
	values.Set("SignatureNonce", randomHex(16))
	values.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	values.Set("AccessKeyId", p.cfg.AccessKeyID)
	values.Set("Signature", aliyunSignature(values, p.cfg.AccessKeySecret))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.Endpoint+"?"+values.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("aliyun sms endpoint returned non-2xx status")
	}
	var payload struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	if payload.Code != "OK" {
		if payload.Message != "" {
			return errors.New(payload.Message)
		}
		return errors.New("aliyun sms send failed")
	}
	return nil
}

// validate checks required Aliyun SMS settings before sending.
// validate 在发送前检查必要的阿里云短信配置。
func (p *AliyunSMSProvider) validate() error {
	if p.cfg.AccessKeyID == "" || p.cfg.AccessKeySecret == "" {
		return errors.New("aliyun sms access key is required")
	}
	if p.cfg.SignName == "" || p.cfg.TemplateCode == "" {
		return errors.New("aliyun sms sign name and template code are required")
	}
	return nil
}

// aliyunSignature signs RPC query parameters with HMAC-SHA1.
// aliyunSignature 使用 HMAC-SHA1 对阿里云 RPC 查询参数签名。
func aliyunSignature(values url.Values, secret string) string {
	canonical := canonicalQuery(values)
	stringToSign := "GET&%2F&" + percentEncode(canonical)
	mac := hmac.New(sha1.New, []byte(secret+"&"))
	_, _ = mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// canonicalQuery returns sorted and percent-encoded query parameters without Signature.
// canonicalQuery 返回排序后的查询参数，排除 Signature 字段。
func canonicalQuery(values url.Values) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		if key == "Signature" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, percentEncode(key)+"="+percentEncode(values.Get(key)))
	}
	return strings.Join(parts, "&")
}

// percentEncode applies Aliyun RPC percent-encoding rules.
// percentEncode 应用阿里云 RPC 签名要求的百分号编码规则。
func percentEncode(value string) string {
	encoded := url.QueryEscape(value)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}

// randomHex creates a random lowercase hex nonce.
// randomHex 创建随机小写十六进制 nonce。
func randomHex(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf)
}
