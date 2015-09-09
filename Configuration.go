package wemvc

type ConnConfig interface {
	GetName() string
	GetType() string
	GetConnString() string
}

type Setting interface {
	GetKey() string
	GetValue() string
}

type Configuration interface {
	GetConnConfig(string) ConnConfig
	GetPort() int
	GetSetting(string) string
}

type connConfig struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	ConnString string `json:"connString"`
}

func (this *connConfig) GetName() string {
	return this.Name
}

func (this *connConfig) GetType() string {
	return this.Type
}

func (this *connConfig) GetConnString() string {
	return this.ConnString
}

type setting struct {
	Key string `json:"key"`
	Value string `json:"value"`
}

func (this *setting)GetKey() string {
	return this.Key
}

func (this *setting)GetValue() string {
	return this.Value
}

type configuration struct {
	Port        int          `json:"port"`
	ConnStrings []connConfig `json:"connStrings"`
	Settings []setting `json:"settings"`
}

func (this *configuration)GetPort() int {
	return this.Port
}

func (this *configuration) GetConnConfig(connName string) ConnConfig {
	for i := 0; i < len(this.ConnStrings); i++ {
		if this.ConnStrings[i].Name == connName{
			return &(this.ConnStrings[i])
		}
	}
	return nil
}

func (this *configuration)GetSetting(key string) string {
	for i := 0; i < len(this.Settings); i++ {
		if this.Settings[i].Key == key{
			return this.Settings[i].Value
		}
	}
	return ""
}

