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
						".terraform/terraform-networking-prod.tfstate",
						`bar = "module1_prod"
env_specific_var = "prodvalue"
foo = "42"
global_var = "globalvalue"
mapvar = <<EOT
{
  entry1 = {
    value1 = testvalue1
    value2 = true
  }
  entry2 = {
    value1 = testvalue2
    value2 = false
  }
}
EOT
myvar = "value for networking"
var_from_env_file = "prod-env"
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
						".terraform/terraform-compute-dev.tfstate",
						`bar = "module2_dev"
env_specific_var = "devvalue"
foo = "42"
global_var = "globalvalue"
mapvar = <<EOT
{
  entry1 = {
    value1 = testvalue1
    value2 = true
  }
  entry2 = {
    value1 = testvalue2
    value2 = false
  }
}
EOT
myvar = "value for compute"
var_from_env_file = "dev-env"
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
						"Backend configuration changed",
					},
					wantErr: true,
				},
			},
		},
		{
			name: "no-vars",
			runs: []testRun{
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=dev", "-m", "testdata/01_networking", "init", "-no-color"},
					want: []string{"Terraform has been successfully initialized!"},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=dev", "-m", "testdata/01_networking", "plan", "-input=false", "-no-color", "-out=plan.out"},
					want: []string{
						"# null_resource.echo will be created",
						"Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config.yaml", "-p", "environment=dev", "-m", "testdata/01_networking", "--no-vars", "apply", "-no-color", "plan.out"},
					want: []string{
						"Apply complete!",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.RemoveAll("testdata/01_networking/.terraform")
				os.RemoveAll("testdata/01_networking/plan.out")
				os.RemoveAll("testdata/02_compute/.terraform")
				os.RemoveAll("testdata/02_compute/plan.out")
				os.Remove("testdata/01_networking/.terraform.lock.hcl")
				os.Remove("testdata/02_compute/.terraform.lock.hcl")
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
