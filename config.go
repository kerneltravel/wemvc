package wemvc

import (
	"strings"

	"github.com/Simbory/wemvc/session"
)

// ConnConfig connection string config section interface
type ConnConfig interface {
	GetName() string
	GetType() string
	GetConnString() string
}

// Configuration global configuration section interface
type Configuration interface {
	GetDefaultUrls() []string
	GetConnConfig(string) ConnConfig
	GetSetting(string) string
	GetMIME(string) string
	GetSessionConfig() *session.ManagerConfig
}

type connConfig struct {
	Name       string `xml:"name,attr"`
	Type       string `xml:"type,attr"`
	ConnString string `xml:"connString,attr"`
}

func (conf *connConfig) GetName() string {
	return conf.Name
}

func (conf *connConfig) GetType() string {
	return conf.Type
}

func (conf *connConfig) GetConnString() string {
	return conf.ConnString
}

type connGroup struct {
	ConfigSource string       `xml:"configSource,attr"`
	ConnStrings  []connConfig `xml:"add"`
}

type setting struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

type settingGroup struct {
	ConfigSource string    `xml:"configSource,attr"`
	Settings     []setting `xml:"add"`
}

type mimeSetting struct {
	FileExe string `xml:"ext,attr"`
	Mime    string `xml:"mime,attr"`
}

type mimeGroup struct {
	ConfigSource string        `xml:"configSource,attr"`
	Mimes        []mimeSetting `xml:"add"`
}

type configuration struct {
	DefaultURL    string                 `xml:"defaultUrl"`
	ConnStrings   connGroup              `xml:"connStrings"`
	Settings      settingGroup           `xml:"settings"`
	Mimes         mimeGroup              `xml:"mime"`
	SessionConfig *session.ManagerConfig `xml:"session"`
	mimeColl      map[string]string
}

func (conf *configuration) GetConnConfig(connName string) ConnConfig {
	for i := 0; i < len(conf.ConnStrings.ConnStrings); i++ {
		if conf.ConnStrings.ConnStrings[i].Name == connName {
			return &(conf.ConnStrings.ConnStrings[i])
		}
	}
	return nil
}

func (conf *configuration) GetSessionConfig() *session.ManagerConfig {
	return conf.SessionConfig
}

func (conf *configuration) GetSetting(key string) string {
	for i := 0; i < len(conf.Settings.Settings); i++ {
		if conf.Settings.Settings[i].Key == key {
			return conf.Settings.Settings[i].Value
		}
	}
	return ""
}

func (conf *configuration) GetMIME(ext string) string {
	if len(ext) < 1 {
		return ""
	}
	if conf.mimeColl == nil {
		conf.mimeColl = make(map[string]string)
		for _, mime := range conf.Mimes.Mimes {
			if len(mime.FileExe) < 1 || len(mime.Mime) < 1 {
				continue
			}
			conf.mimeColl[strings.ToLower(mime.FileExe)] = mime.Mime
		}
	}
	return conf.mimeColl[strings.ToLower(ext)]
}

func (conf *configuration) GetDefaultUrls() []string {
	if len(conf.DefaultURL) < 1 {
		return nil
	}
	return strings.Split(conf.DefaultURL, ";")
}
