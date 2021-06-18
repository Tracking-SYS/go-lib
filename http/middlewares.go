package http

import (
	"net/http"
	"strings"

	"github.com/lk153/go-lib/reflection"
)

type userMetadata struct {
	CustomerID int64  `query_string:"customer_id"`
	TrackityID string `query_string:"trackity_id"`
}

type DeviceInfo struct {
	// android, ios, web, mweb
	PlatformName string `json:"platform_name,omitempty"`
	// current version x.y.z (app only)
	AppVersion string `json:"app_version,omitempty"`
	// provided or fallback from useragent
	OsName string `json:"os_name,omitempty"`
	// provided or fallback from useragent
	OsVersion string `json:"os_version,omitempty"`
	// browser (automatic extract from user-agent)
	BrowserName string `json:"browser_name,omitempty"`
	// version of browser (automatic extract from user-agent)
	BrowserVersion string `json:"browser_version,omitempty"`
	// engine of browser (automatic extract from user-agent)
	BrowserEngine string `json:"browser_engine,omitempty"`
	// phone manufactor: samsung, apple ...
	DeviceManufacturer string `json:"device_manufacturer,omitempty"`
	// which network user are using (mobile network: viettel, vinaphone)
	NetworkCarrier string `json:"network_carrier,omitempty"`
	// google, facebook bot
	IsBot bool `json:"is_bot,omitempty"`
}

type RequestInfo struct {
	userMetadata userMetadata
	DeviceInfo   *DeviceInfo
}

func ExtractUserMetaFromRequest(req *http.Request) interface{} {
	query := req.URL.Query()
	var queryParams userMetadata
	_ = DecodeQueryParams(query, &queryParams)

	requestInfo := new(RequestInfo)
	requestInfo.userMetadata = queryParams

	uaRaw := req.Header.Get(HeaderUserAgent)
	if uaRaw != "" {
		requestInfo.DeviceInfo = UserAgentParser(uaRaw)
	} else {
		requestInfo.DeviceInfo = &DeviceInfo{}
	}

	// this is special case for platform
	if query.Get("is_mweb") == "1" {
		requestInfo.DeviceInfo.PlatformName = PlatformMWeb
	}

	if requestInfo.DeviceInfo.PlatformName == "" {
		requestInfo.DeviceInfo.PlatformName = PlatformWeb
	}

	return requestInfo
}

// Rules priority order:
// 1.header Authorization
// 2.header X-Access-Token
// 3.cookie tracking-access-token
func GetTokenFromReq(req *http.Request) string {
	return reflection.Coalesce(
		strings.TrimSpace(strings.TrimPrefix(req.Header.Get(HeaderAuthorization), "Bearer")),
		strings.TrimSpace(req.Header.Get(HeaderXAccessToken)),
		GetCookieVal(req, CookieAccessToken),
	).(string)
}

func GetCookieVal(req *http.Request, key string) string {
	if cookie, err := req.Cookie(key); err == nil && cookie.Value != "" {
		return strings.TrimSpace(cookie.Value)
	}

	return ""
}
