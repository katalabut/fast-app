package source

import (
	"strings"

	"github.com/spf13/viper"
)

const EnvSourceName = "env"

type Env struct {
	prefix string
}

func NewEnv(prefix string) *Env {
	return &Env{prefix: prefix}
}

func (e *Env) Name() string {
	return EnvSourceName
}

func (e *Env) Load(v *viper.Viper) error {
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AllowEmptyEnv(true)
	v.SetEnvPrefix(e.prefix)
	v.AutomaticEnv()

	return nil
}
