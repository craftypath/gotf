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
)

func TestLoad(t *testing.T) {
	got, err := Load("testdata/test_config.yaml", "testdata/testmodule", map[string]string{
		"param": "paramvalue",
		"env":   "prod",
	})
	assert.NoError(t, err)

	assert.Equal(t, []string{
		"../testmodule/test1-prod.tfvars",
		"../testmodule/test2-prod.tfvars",
	}, got.VarFiles)
	assert.Equal(t, map[string]string{
		"foo":          "foovalue",
		"templatedVar": "paramvalue",
		"mapvar":       "{\n  value1 = \"testvalue\"\n  value2 = true\n}",
	}, got.Vars)
	assert.Equal(t, map[string]string{
		"BAR":           "barvalue",
		"TEMPLATED_ENV": "paramvalue",
	}, got.Envs)
	assert.Equal(t, map[string]string{
		"backend_key":                  "be_key_foovalue_barvalue",
		"backend_storage_account_name": "be_storage_account_name_foovalue_barvalue",
		"backend_resource_group_name":  "be_resource_group_name_foovalue_barvalue",
		"backend_container_name":       "be_container_name_foovalue_barvalue",
	}, got.BackendConfigs)
}
