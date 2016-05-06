package session

type ManagerConfig struct {
	ManagerName     string `xml:"manager,attr"`
	CookieName      string `xml:"cookieName,attr"`
	EnableSetCookie bool   `xml:"enableSetCookie,attr"`
	Gclifetime      int64  `xml:"gclifetime,attr"`
	Maxlifetime     int64  `xml:"maxLifetime,attr"`
	Secure          bool   `xml:"secure,attr"`
	CookieLifeTime  int    `xml:"cookieLifeTime,attr"`
	ProviderConfig  string `xml:"providerConfig,attr"`
	Domain          string `xml:"domain,attr"`
	SessionIDLength int64  `xml:"sessionIDLength,attr"`
}