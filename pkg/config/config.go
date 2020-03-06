// Copyright The gotf Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v2"
)

type Config struct {
	TerraformVersion string            `yaml:"terraformVersion"`
	VarFiles         []string          `yaml:"varFiles"`
	Vars             map[string]string `yaml:"vars"`
	Envs             map[string]string `yaml:"envs"`
	BackendConfigs   map[string]string `yaml:"backendConfigs"`
}

func Load(configFile string, moduleDir string, params map[string]string) (*Config, error) {
	log.Println("Loading config file:", configFile)
	cfgData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not load config file %q: %w", configFile, err)
	}

	cfg, err := load(cfgData)
	if err != nil {
		return nil, fmt.Errorf("could not load config file %q: %w", configFile, err)
	}

	templatingInput := map[string]interface{}{
		"Params": params,
	}

	log.Println("Processing varFiles...")
	cfgFileDir := filepath.Dir(configFile)
	cfgFileDirRelativeToModuleDir, err := filepath.Rel(moduleDir, cfgFileDir)
	if err != nil {
		return nil, err
	}
	for i, f := range cfg.VarFiles {
		sb := strings.Builder{}
		err := renderTemplate(&sb, templatingInput, f)
		if err != nil {
			return nil, err
		}
		varFile := sb.String()
		if !filepath.IsAbs(f) {
			varFile = filepath.Join(cfgFileDirRelativeToModuleDir, varFile)
		}
		cfg.VarFiles[i] = varFile
	}

	log.Println("Processing vars...")
	for key, value := range cfg.Vars {
		sb := strings.Builder{}
		if err := renderTemplate(&sb, templatingInput, value); err != nil {
			return nil, err
		}
		cfg.Vars[key] = sb.String()
	}

	log.Println("Processing envs...")
	for key, value := range cfg.Envs {
		sb := strings.Builder{}
		if err := renderTemplate(&sb, templatingInput, value); err != nil {
			return nil, err
		}
		cfg.Envs[key] = sb.String()
	}

	templatingInput = map[string]interface{}{
		"Vars":   cfg.Vars,
		"Envs":   cfg.Envs,
		"Params": params,
	}

	log.Println("Proessing backendConfigs...")
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
	tpl := template.New("gotpl").Funcs(sprig.TxtFuncMap()).Option("missingkey=error")
	tpl, err := tpl.Parse(text)
	if err != nil {
		return err
	}
	return tpl.Execute(wr, data)
}
