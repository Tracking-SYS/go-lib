package http

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lk153/go-lib/log"
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
	queryParams.CustomerID = getCustomerID(req, queryParams)
	getDeviceInfo(req, requestInfo)
	getPlatformName(query, requestInfo)

	return requestInfo
}

func getDeviceInfo(req *http.Request, requestInfo *RequestInfo) {
	uaRaw := req.Header.Get(HeaderUserAgent)
	if uaRaw != "" {
		requestInfo.DeviceInfo = UserAgentParser(uaRaw)
	} else {
		requestInfo.DeviceInfo = &DeviceInfo{}
	}
}

func getPlatformName(query url.Values, requestInfo *RequestInfo) {
	if query.Get("is_mweb") == "1" {
		requestInfo.DeviceInfo.PlatformName = PlatformMWeb
	}

	if requestInfo.DeviceInfo.PlatformName == "" {
		requestInfo.DeviceInfo.PlatformName = PlatformWeb
	}
}

func getCustomerID(req *http.Request, queryMeta userMetadata) int64 {
	token := GetTokenFromReq(req)
	if token == "" {
		return 0
	}

	customerID, err := decodeTokenUnVerify(token)
	if err != nil {
		log.Log.Info("decode token got error", "token", token, "error", err)
		return 0
	}

	customerID = reflection.Coalesce(
		customerID,
		queryMeta.CustomerID,
	).(int64)

	return customerID
}

func decodeTokenUnVerify(token string) (int64, error) {
	t, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return 0, err
	}

	customerIDRaw, ok := t.Claims.(jwt.MapClaims)["customer_id"]
	if !ok {
		return 0, fmt.Errorf("invalid customer_id field in token: %w", err)
	}

	customerID, err := reflection.ParseInt64(customerIDRaw)
	if err != nil {
		return 0, fmt.Errorf("invalid customer_id data in token. Expected int datatype: %w", err)
	}
	return customerID, nil
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
