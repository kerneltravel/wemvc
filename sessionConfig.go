package wemvc

// SessionConfig the session config struct
type SessionConfig struct {
	ManagerName     string `xml:"manager,attr"`
	CookieName      string `xml:"cookieName,attr"`
	EnableSetCookie bool   `xml:"enableSetCookie,attr"`
	GcLifetime      int64  `xml:"gclifetime,attr"`
	MaxLifetime     int64  `xml:"maxLifetime,attr"`
	Secure          bool   `xml:"secure,attr"`
	HttpOnly        bool   `xml:"httpOnly,attr"`
	CookieLifeTime  int    `xml:"cookieLifeTime,attr"`
	ProviderConfig  string `xml:"providerConfig,attr"`
	Domain          string `xml:"domain,attr"`
	SessionIDLength int64  `xml:"sessionIDLength,attr"`
}
