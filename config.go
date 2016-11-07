package wemvc

import (
	"strings"
	"fmt"
)

// Configuration the global config interface
type Configuration interface {
	GetDefaultUrls() []string
	GetConnConfig(string) (typeName string, connString string)
	GetSetting(string) string
}

type connSetting struct {
	typeName   string
	connString string
}

type config struct {
	DefaultURL  string `xml:"defaultUrl"`
	ConnStrings struct {
		List []struct {
			Name       string `xml:"name,attr"`
			Type       string `xml:"type,attr"`
			ConnString string `xml:"connString,attr"`
		} `xml:"add"`
	} `xml:"connStrings"`
	Settings struct {
		List []struct {
			Key   string `xml:"key,attr"`
			Value string `xml:"value,attr"`
		} `xml:"add"`
	} `xml:"settings"`
	SessionConfig *SessionConfig `xml:"session"`
	settingMap    map[string]string
	connMap       map[string]*connSetting
	defaultUrls   []string
}

func (conf *config) loadFile(file string) error {
	conf.settingMap = make(map[string]string)
	conf.connMap = make(map[string]*connSetting)
	err := file2Xml(file, conf)
	if err != nil {
		return err
	}
	if len(conf.Settings.List) > 0 {
		for _, s := range conf.Settings.List {
			if len(s.Key) < 1 {
				continue
			}
			conf.settingMap[s.Key] = s.Value
		}
	}
	if len(conf.ConnStrings.List) > 0 {
		for _, conn := range conf.ConnStrings.List {
			if len(conn.Name) < 1 {
				continue
			}
			conf.connMap[conn.Name] = &connSetting{typeName: conn.Type, connString: conn.ConnString}
		}
	}
	if len(conf.DefaultURL) > 0 {
		splits := strings.Split(conf.DefaultURL, ";,")
		for _, s := range splits {
			if len(s) < 1 {
				continue
			}
			conf.defaultUrls = append(conf.defaultUrls, s)
		}
	}
	if len(conf.defaultUrls) < 1 {
		conf.defaultUrls = []string{"index.html", "index.htm"}
	}
	if conf.SessionConfig == nil {
		conf.SessionConfig = &SessionConfig{}
	}
	if len(conf.SessionConfig.ManagerName) < 1 {
		conf.SessionConfig.ManagerName = "memory"
	}
	if conf.SessionConfig.GcLifetime == 0 {
		conf.SessionConfig.GcLifetime = 3600
	}
	if conf.SessionConfig.MaxLifetime == 0 {
		conf.SessionConfig.MaxLifetime = 3600
	}
	return nil
}

func (conf *config) GetConnConfig(connName string) (string, string) {
	conn, ok := conf.connMap[connName]
	if ok {
		return conn.typeName, conn.connString
	}
	return "", ""
}

func (conf *config) getSessionConfig() *SessionConfig {
	return conf.SessionConfig
}

func (conf *config) GetSetting(key string) string {
	v, ok := conf.settingMap[key]
	if ok {
		return v
	}
	return ""
}

func (conf *config) GetDefaultUrls() []string {
	return conf.defaultUrls
}

func newConfig(strPath string) (*config, error) {
	if !IsFile(strPath) {
		return nil, fmt.Errorf("Canot fild the config file: %s", strPath)
	}
	conf := &config{}
	err := conf.loadFile(strPath)
	if err != nil {
		return nil, err
	}
	return conf, nil
}