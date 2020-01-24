// Copyright The gotf Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

type Config struct {
	VarsFiles      []string          `yaml:"varsFiles"`
	Vars           map[string]string `yaml:"vars"`
	Envs           map[string]string `yaml:"envs"`
	BackendConfigs map[string]string `yaml:"backendConfigs"`
}

func Load(configFile string) (*Config, error) {
	cfgData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	cfg, err := load(cfgData)
	if err != nil {
		return nil, err
	}

	cfgFileDir := filepath.Dir(configFile)
	for i, f := range cfg.VarsFiles {
		if !filepath.IsAbs(f) {
			varFile := filepath.Join(cfgFileDir, f)
			cfg.VarsFiles[i] = varFile
		}
	}
	templatingInput := map[string]interface{}{
		"Vars": cfg.Vars,
		"Envs": cfg.Envs,
	}

	for key, value := range cfg.BackendConfigs {
		sb := strings.Builder{}
		if err := renderTemplate(&sb, templatingInput, value); err != nil {
			return nil, err
		}
		cfg.BackendConfigs[key] = sb.String()
	}
	return cfg, nil
}

func load(cfgData []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(cfgData, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func renderTemplate(wr io.Writer, data map[string]interface{}, text string) error {
	tpl := template.New("gotpl")
	tpl, err := tpl.Parse(text)
	if err != nil {
		return err
	}
	return tpl.Execute(wr, data)
}
