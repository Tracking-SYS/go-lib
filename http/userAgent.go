package http

import (
	"strings"

	"github.com/ua-parser/uap-go/uaparser"
	"gopkg.in/yaml.v2"
)

const (
	botFamily = "spider"

	PlatformWeb     = "web"
	PlatformMWeb    = "mweb"
	PlatformAndroid = "android"
	PlatformIOS     = "ios"
	PlatformUnKnown = "unknown"
)

//go:embed "predefined.yaml"
var predefinedBuf []byte

//go:embed "tiki.yaml"
var customBuf []byte

var parser = func() *uaparser.Parser {
	var originalDef uaparser.RegexesDefinitions
	var custom uaparser.RegexesDefinitions
	if err := yaml.Unmarshal(predefinedBuf, &originalDef); err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(customBuf, &custom); err != nil {
		panic(err)
	}
	originalDef.UA = append(custom.UA, originalDef.UA...)
	originalDef.OS = append(custom.OS, originalDef.OS...)
	originalDef.Device = append(custom.Device, originalDef.Device...)

	buf, err := yaml.Marshal(&originalDef)
	if err != nil {
		panic(err)
	}
	p, err := uaparser.NewFromBytes(buf)
	if err != nil {
		panic(err)
	}
	return p
}()

func UserAgentParser(val string) *DeviceInfo {
	d := &DeviceInfo{}
	if val == "" {
		d.PlatformName = PlatformUnKnown
		return d
	}
	ua := parser.Parse(strings.TrimSpace(val))

	// filter bot
	if strings.EqualFold(ua.Device.Family, botFamily) ||
		strings.Contains(strings.ToLower(ua.UserAgent.Family), botFamily) ||
		strings.Contains(strings.ToLower(val), "bot") {
		d.IsBot = true
		return d
	}
	family := strings.ToLower(ua.UserAgent.Family)
	osFamily := strings.ToLower(ua.Os.Family)
	d.PlatformName = osFamily
	d.AppVersion = uaVersion(ua.UserAgent)

	// generic android & ios
	if family == PlatformAndroid || family == PlatformIOS {
		d.PlatformName = family
		return d
	}

	d.BrowserName = family
	d.BrowserVersion = uaVersion(ua.UserAgent)
	d.OsName = strings.ToLower(ua.Os.Family)
	// mweb
	if strings.Contains(strings.ToLower(family), "mobile") || osFamily == PlatformAndroid || osFamily == PlatformIOS || family == "facebook" {
		d.PlatformName = PlatformMWeb
	} else {
		d.PlatformName = PlatformWeb
	}

	return d
}

func uaVersion(ua *uaparser.UserAgent) string {
	return ua.ToVersionString()
}
