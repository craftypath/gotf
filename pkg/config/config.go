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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

// All maps reresenting YAML dicts are of type map[string]interface{} so Sprig collection
// functions can be used because they expect this type.
type fileConfig struct {
	TerraformVersion      string                            `yaml:"terraformVersion"`
	RequiredParams        map[string][]string               `yaml:"requiredParams"`
	Params                map[string]interface{}            `yaml:"params"`
	GlobalVarFiles        []string                          `yaml:"globalVarFiles"`
	ModuleVarFiles        map[string][]string               `yaml:"moduleVarFiles"`
	GlobalVars            map[string]interface{}            `yaml:"globalVars"`
	ModuleVars            map[string]map[string]interface{} `yaml:"moduleVars"`
	Envs                  map[string]string                 `yaml:"envs"`
	VarsFromEnvFiles      []string                          `yaml:"varsFromEnvFiles"`
	BackendConfigs        map[string]string                 `yaml:"backendConfigs"`
	IgnoreMissingVarFiles bool                              `yaml:"ignoreMissingVarFiles"`
}

type Config struct {
	TerraformVersion string
	VarFiles         []string
	Vars             map[string]string
	Envs             map[string]string
	BackendConfigs   map[string]string
}

const moduleDirParamName = "moduleDir"

func Load(configFile string, modulePath string, cliParams map[string]string) (*Config, error) {
	log.Println("Loading config file:", configFile)
	cfgData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	fileCfg, err := load(cfgData)
	if err != nil {
		return nil, err
	}

	if err := checkRequiredParams(fileCfg, cliParams); err != nil {
		return nil, err
	}

	params := make(map[string]interface{})
	if err := appendInterfaceParams(params, fileCfg.Params); err != nil {
		return nil, err
	}
	if err := appendStringParams(params, cliParams); err != nil {
		return nil, err
	}

	// we need the absolute path here in case "." was specified as relative path
	abs, err := filepath.Abs(modulePath)
	if err != nil {
		return nil, err
	}
	params[moduleDirParamName] = filepath.Base(abs)

	log.Println("Processing var files...")
	cfgFileDir := filepath.Dir(configFile)

	cfg := &Config{
		TerraformVersion: fileCfg.TerraformVersion,
		VarFiles:         []string{},
		Vars:             make(map[string]string),
		Envs:             make(map[string]string),
		BackendConfigs:   make(map[string]string),
	}

	for _, f := range fileCfg.GlobalVarFiles {
		varFilePath, err := computeModuleRelativePath(f, params, cfgFileDir, modulePath)
		if err != nil {
			return nil, err
		}
		if err := maybeAppendValFile(cfg, fileCfg.IgnoreMissingVarFiles, varFilePath, modulePath); err != nil {
			return nil, err
		}
	}

	log.Println("Processing module var files...")
	moduleDir := params[moduleDirParamName].(string)
	for _, f := range fileCfg.ModuleVarFiles[moduleDir] {
		varFilePath, err := computeModuleRelativePath(f, params, cfgFileDir, modulePath)
		if err != nil {
			return nil, err
		}
		if err := maybeAppendValFile(cfg, fileCfg.IgnoreMissingVarFiles, varFilePath, modulePath); err != nil {
			return nil, err
		}
	}

	log.Println("Processing vars from env files...")
	for _, f := range fileCfg.VarsFromEnvFiles {
		path, err := computeModuleRelativePath(f, params, cfgFileDir, modulePath)
		if err != nil {
			return nil, err
		}
		if err := maybeAppendVarsFromEnvFile(cfg, fileCfg.IgnoreMissingVarFiles, path, modulePath); err != nil {
			return nil, err
		}
	}

	log.Println("Processing global vars...")
	for key, value := range fileCfg.GlobalVars {
		result, err := computeValue(value, params)
		if err != nil {
			return nil, err
		}
		cfg.Vars[key] = result
	}

	log.Println("Processing module vars...")
	for key, value := range fileCfg.ModuleVars[moduleDir] {
		result, err := computeValue(value, params)
		if err != nil {
			return nil, err
		}
		cfg.Vars[key] = result
	}

	log.Println("Processing envs...")
	for key, value := range fileCfg.Envs {
		result, err := computeValue(value, params)
		if err != nil {
			return nil, err
		}
		cfg.Envs[key] = result
	}

	templatingInput := map[string]interface{}{
		"Vars":   cfg.Vars,
		"Envs":   cfg.Envs,
		"Params": params,
	}

	log.Println("Processing backend configs...")
	for key, value := range fileCfg.BackendConfigs {
		result, err := renderTemplate(templatingInput, value)
		if err != nil {
			return nil, err
		}
		cfg.BackendConfigs[key] = result
	}

	return cfg, nil
}

func maybeAppendValFile(cfg *Config, ignoreMissingVarFiles bool, varFilePath string, modulePath string) error {
	if ignoreMissingVarFiles {
		var path string
		if filepath.IsAbs(varFilePath) {
			path = varFilePath
		} else {
			path = filepath.Join(modulePath, varFilePath)
		}
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				log.Println(fmt.Sprintf("File %s does not exist. Ignoring it.", path))
				return nil
			}
		}
	}

	cfg.VarFiles = append(cfg.VarFiles, varFilePath)
	return nil
}

func maybeAppendVarsFromEnvFile(cfg *Config, IgnoreMissingVarFiles bool, varFilePath string, modulePath string) error {
	var path string
	if filepath.IsAbs(varFilePath) {
		path = varFilePath
	} else {
		path = filepath.Join(modulePath, varFilePath)
	}

	if IgnoreMissingVarFiles {
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				log.Println(fmt.Sprintf("File %s does not exist. Ignoring it.", path))
				return nil
			}
		}
	}

	envs, err := godotenv.Read(path)
	if err != nil {
		return err
	}
	for key, value := range envs {
		cfg.Vars[strings.ToLower(key)] = value
	}
	return nil
}

func checkRequiredParams(fileCfg *fileConfig, cliParams map[string]string) error {
	for k, v := range fileCfg.RequiredParams {
		value, ok := cliParams[k]
		if !ok {
			return fmt.Errorf("required parameter %q must be specified", k)
		}
		if len(v) > 0 {
			var hasAllowedValue bool
			for _, allowed := range v {
				if value == allowed {
					hasAllowedValue = true
					break
				}
			}
			if !hasAllowedValue {
				return fmt.Errorf("value for required parameter %q must be one of %v", k, v)
			}
		}
	}
	return nil
}

func load(cfgData []byte) (*fileConfig, error) {
	var cfg fileConfig
	if err := yaml.Unmarshal(cfgData, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func renderTemplate(data map[string]interface{}, tmpl string) (string, error) {
	wr := strings.Builder{}
	tpl := template.New("gotpl").Funcs(sprig.HermeticTxtFuncMap()).Option("missingkey=error")
	tpl, err := tpl.Parse(tmpl)
	if err != nil {
		return "", err
	}
	if err := tpl.Execute(&wr, data); err != nil {
		return "", err
	}
	return wr.String(), nil
}

func computeModuleRelativePath(pathTemplate string, params map[string]interface{}, cfgFileDir string, modulePath string) (string, error) {
	templatingInput := map[string]interface{}{
		"Params": params,
	}
	path, err := renderTemplate(templatingInput, pathTemplate)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(path) {
		varFilePath := filepath.Join(cfgFileDir, path)
		if varFilePath, err = filepath.Rel(modulePath, varFilePath); err != nil {
			return "", err
		}
		return varFilePath, nil
	}
	return path, nil
}

func computeValue(valueTemplate interface{}, params map[string]interface{}) (string, error) {
	if tmpl, ok := valueTemplate.(string); ok {
		templatingInput := map[string]interface{}{
			"Params": params,
		}
		return renderTemplate(templatingInput, tmpl)
	}
	return fmt.Sprint(valueTemplate), nil
}

func appendStringParams(dst map[string]interface{}, src map[string]string) error {
	for k, v := range src {
		if err := checkForModuleDirParam(k); err != nil {
			return err
		}
		dst[k] = v
	}
	return nil
}

func appendInterfaceParams(dst map[string]interface{}, src map[string]interface{}) error {
	for k, v := range src {
		if err := checkForModuleDirParam(k); err != nil {
			return err
		}
		dst[k] = v
	}
	return nil
}

func checkForModuleDirParam(key string) error {
	if key == moduleDirParamName {
		return fmt.Errorf("param %q is reserved and set automatically", moduleDirParamName)
	}
	return nil
}
