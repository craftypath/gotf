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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	type args struct {
		configFile string
		moduleDir  string
		params     map[string]string
	}
	tests := []struct {
		name       string
		args       args
		want       *Config
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "Load dev config testmodule1",
			args: struct {
				configFile string
				moduleDir  string
				params     map[string]string
			}{
				configFile: "testdata/test-config.yaml",
				moduleDir:  "testmodule1",
				params: map[string]string{
					"environment": "dev",
				},
			},
			want: &Config{
				TerraformVersion: "0.12.24",
				VarFiles: []string{
					"../testdata/global.tfvars",
					"../testdata/global-dev.tfvars",
					"../testdata/testmodule1/test1-dev.tfvars",
					"../testdata/testmodule1/test2-dev.tfvars",
				},
				Vars: map[string]string{
					"foo":          "foovalue",
					"templatedVar": "paramvalue",
					"mapvar":       "{\n  value1 = \"testvalue\"\n  value2 = true\n}",
					"moduleVar1":   "testmodule1_value1",
					"moduleVar2":   "testmodule1_value2",
				},
				Envs: map[string]string{
					"BAR":           "barvalue",
					"TEMPLATED_ENV": "paramvalue",
				},
				BackendConfigs: map[string]string{
					"key":                  "testmodule1",
					"storage_account_name": "mytfstateaccountdev",
					"resource_group_name":  "mytfstate-dev",
					"container_name":       "mytfstate-dev",
				},
			},
		},
		{
			name: "Load dev config testmodule2",
			args: struct {
				configFile string
				moduleDir  string
				params     map[string]string
			}{
				configFile: "testdata/test-config.yaml",
				moduleDir:  "testmodule2",
				params: map[string]string{
					"environment": "dev",
				},
			},
			want: &Config{
				TerraformVersion: "0.12.24",
				VarFiles: []string{
					"../testdata/global.tfvars",
					"../testdata/global-dev.tfvars",
					"../testdata/testmodule2/test1-dev.tfvars",
					"../testdata/testmodule2/test2-dev.tfvars",
				},
				Vars: map[string]string{
					"foo":          "foovalue",
					"templatedVar": "paramvalue",
					"mapvar":       "{\n  value1 = \"testvalue\"\n  value2 = true\n}",
					"moduleVar1":   "testmodule2_value1",
					"moduleVar2":   "testmodule2_value2",
				},
				Envs: map[string]string{
					"BAR":           "barvalue",
					"TEMPLATED_ENV": "paramvalue",
				},
				BackendConfigs: map[string]string{
					"key":                  "testmodule2",
					"storage_account_name": "mytfstateaccountdev",
					"resource_group_name":  "mytfstate-dev",
					"container_name":       "mytfstate-dev",
				},
			},
		},
		{
			name: "Load prod config testmodule1",
			args: struct {
				configFile string
				moduleDir  string
				params     map[string]string
			}{
				configFile: "testdata/test-config.yaml",
				moduleDir:  "testmodule1",
				params: map[string]string{
					"environment": "prod",
				},
			},
			want: &Config{
				TerraformVersion: "0.12.24",
				VarFiles: []string{
					"../testdata/global.tfvars",
					"../testdata/global-prod.tfvars",
					"../testdata/testmodule1/test1-prod.tfvars",
					"../testdata/testmodule1/test2-prod.tfvars",
				},
				Vars: map[string]string{
					"foo":          "foovalue",
					"templatedVar": "paramvalue",
					"mapvar":       "{\n  value1 = \"testvalue\"\n  value2 = true\n}",
					"moduleVar1":   "testmodule1_value1",
					"moduleVar2":   "testmodule1_value2",
				},
				Envs: map[string]string{
					"BAR":           "barvalue",
					"TEMPLATED_ENV": "paramvalue",
				},
				BackendConfigs: map[string]string{
					"key":                  "testmodule1",
					"storage_account_name": "mytfstateaccountprod",
					"resource_group_name":  "mytfstate-prod",
					"container_name":       "mytfstate-prod",
				},
			},
		},
		{
			name: "Load prod config testmodule2",
			args: struct {
				configFile string
				moduleDir  string
				params     map[string]string
			}{
				configFile: "testdata/test-config.yaml",
				moduleDir:  "testmodule2",
				params: map[string]string{
					"environment": "prod",
				},
			},
			want: &Config{
				TerraformVersion: "0.12.24",
				VarFiles: []string{
					"../testdata/global.tfvars",
					"../testdata/global-prod.tfvars",
					"../testdata/testmodule2/test1-prod.tfvars",
					"../testdata/testmodule2/test2-prod.tfvars",
				},
				Vars: map[string]string{
					"foo":          "foovalue",
					"templatedVar": "paramvalue",
					"mapvar":       "{\n  value1 = \"testvalue\"\n  value2 = true\n}",
					"moduleVar1":   "testmodule2_value1",
					"moduleVar2":   "testmodule2_value2",
				},
				Envs: map[string]string{
					"BAR":           "barvalue",
					"TEMPLATED_ENV": "paramvalue",
				},
				BackendConfigs: map[string]string{
					"key":                  "testmodule2",
					"storage_account_name": "mytfstateaccountprod",
					"resource_group_name":  "mytfstate-prod",
					"container_name":       "mytfstate-prod",
				},
			},
		},
		{
			name: "Missing required param",
			args: struct {
				configFile string
				moduleDir  string
				params     map[string]string
			}{
				configFile: "testdata/test-config.yaml",
				moduleDir:  "testmodule1",
			},
			wantErr:    true,
			wantErrMsg: `required parameter "environment" must be specified`,
		},
		{
			name: "Invalid required param value",
			args: struct {
				configFile string
				moduleDir  string
				params     map[string]string
			}{
				configFile: "testdata/test-config.yaml",
				moduleDir:  "testmodule1",
				params: map[string]string{
					"environment": "invalid",
				},
			},
			wantErr:    true,
			wantErrMsg: `value for required parameter "environment" must be one of [dev prod]`,
		},
		{
			name: "Module dir explicitly set",
			args: struct {
				configFile string
				moduleDir  string
				params     map[string]string
			}{
				configFile: "testdata/test-config.yaml",
				moduleDir:  "testmodule1",
				params: map[string]string{
					"environment": "dev",
					"moduleDir":   "dummy",
				},
			},
			wantErr:    true,
			wantErrMsg: `param "moduleDir" is reserved and set automatically`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(tt.args.configFile, tt.args.moduleDir, tt.args.params)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
