package config

type Ge2o struct {
	// ApiSecret 接口本地密钥
	ApiSecret string `yaml:"api-secret"`
}

func (g *Ge2o) Init() error {
	return nil
}
