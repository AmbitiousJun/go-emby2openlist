package config

type Ge2o struct {
	// ApiSecret 接口本地密钥
	ApiSecret string `yaml:"api-secret"`

	// Web web 平台配置
	Web *Web `yaml:"web"`
}

func (g *Ge2o) Init() error {

	if g.Web == nil {
		g.Web = new(Web)
	}

	return nil
}

type Web struct {

	// Disable 是否禁用 web 平台
	Disable bool `yaml:"disable"`

	// DisableEmbyBtn 是否禁用 emby 快速进入 web 平台的辅助按钮
	DisableEmbyBtn bool `yaml:"disable-emby-btn"`
}

// IsEnabled web 平台是否启用
func (w *Web) IsEnabled() bool {
	return !w.Disable
}

// IsEmbyBtnEnabled emby 辅助按钮是否启用
func (w *Web) IsEmbyBtnEnabled() bool {
	return !w.DisableEmbyBtn
}
