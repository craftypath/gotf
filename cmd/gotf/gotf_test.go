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

package gotf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	type testRun struct {
		args    []string
		want    []string
		wantErr bool
	}

	tests := []struct {
		name string
		runs []testRun
	}{
		{
			name: "networking module prod",
			runs: []testRun{
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=prod", "-m", "testdata/01_networking", "init", "-no-color"},
					want: []string{"Terraform has been successfully initialized!"},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=prod", "-m", "testdata/01_networking", "plan", "-input=false", "-no-color"},
					want: []string{
						"# null_resource.echo will be created",
						"Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=prod", "-m", "testdata/01_networking", "apply", "-auto-approve", "-no-color"},
					want: []string{
						"State path: .terraform/terraform-networking-prod.tfstate",
						`bar = module1_prod
envSpecificVar = prodvalue
foo = 42
globalVar = globalvalue
mapvar = {
  entry1 = {
    value1 = testvalue1
    value2 = true
  }
  entry2 = {
    value1 = testvalue2
    value2 = false
  }
}
myvar = value for networking
`},
				},
			},
		},
		{
			name: "compute module dev",
			runs: []testRun{
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=dev", "-m", "testdata/02_compute", "init", "-no-color"},
					want: []string{"Terraform has been successfully initialized!"},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=dev", "-m", "testdata/02_compute", "plan", "-input=false", "-no-color"},
					want: []string{
						"# null_resource.echo will be created",
						"Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=dev", "-m", "testdata/02_compute", "apply", "-auto-approve", "-no-color"},
					want: []string{
						"State path: .terraform/terraform-compute-dev.tfstate",
						`bar = module2_dev
envSpecificVar = devvalue
foo = 42
globalVar = globalvalue
mapvar = {
  entry1 = {
    value1 = testvalue1
    value2 = true
  }
  entry2 = {
    value1 = testvalue2
    value2 = false
  }
}
myvar = value for compute
`},
				},
			},
		},
		{
			name: "backend check",
			runs: []testRun{
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=prod", "-m", "testdata/01_networking", "init", "-no-color"},
					want: []string{"Terraform has been successfully initialized!"},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=prod", "-m", "testdata/01_networking", "plan", "-input=false", "-no-color"},
					want: []string{
						"# null_resource.echo will be created",
						"Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=dev", "-m", "testdata/01_networking", "plan", "-input=false", "-no-color"},
					want: []string{
						"configured backend does not match current environment",
						"path: got=.terraform/terraform-networking-prod.tfstate, want=.terraform/terraform-networking-dev.tfstate",
						"Run terraform init -reconfigure!",
					},
					wantErr: true,
				},
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=dev", "-m", "testdata/01_networking", "init", "-reconfigure"},
					want: []string{
						"Terraform has been successfully initialized!",
					},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=dev", "-m", "testdata/01_networking", "plan", "-input=false", "-no-color"},
					want: []string{
						"# null_resource.echo will be created",
						"Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
			},
		},
		{
			name: "skip backend check",
			runs: []testRun{
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=prod", "-m", "testdata/01_networking", "init", "-no-color"},
					want: []string{"Terraform has been successfully initialized!"},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=prod", "-m", "testdata/01_networking", "plan", "-input=false", "-no-color"},
					want: []string{
						"# null_resource.echo will be created",
						"Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
				{
					args: []string{"-d", "--skip-backend-check", "-c", "testdata/test-config.yaml", "-p", "environment=dev", "-m", "testdata/01_networking", "init", "-no-color"},
					want: []string{
						"Backend configuration changed!",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := ioutil.TempDir("testdata", "gotf")
			t.Cleanup(func() { os.RemoveAll(tempDir) })

			panicOnError(err)
			panicOnError(os.Setenv("XDG_CACHE_HOME", filepath.Join(tempDir, "tfcache")))

			t.Cleanup(func() {
				os.RemoveAll("testdata/01_networking/.terraform")
				os.RemoveAll("testdata/02_compute/.terraform")
				os.Remove("testdata/01_networking/terraform.tfstate")
				os.Remove("testdata/02_compute/terraform.tfstate")
				os.Remove("testdata/01_networking/.terraform.tfstate.lock.info")
				os.Remove("testdata/02_compute/.terraform.tfstate.lock.info")
			})

			for _, run := range tt.runs {
				got, err := runGotf(run.args)
				if run.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				for _, want := range run.want {
					assert.Contains(t, got, want)
				}
			}
		})
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func runGotf(args []string) (string, error) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	command := newGotfCommand()
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	command.SetArgs(args)
	err := command.Execute()
	w.Close()
	bytes, _ := ioutil.ReadAll(r)
	output := string(bytes)
	println(output)
	return output, err
}
