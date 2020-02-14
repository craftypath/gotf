package gotf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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
			name: "happy path",
			runs: []testRun{
				{
					args: []string{"-d", "-c", "testdata/test-config-prod.yaml", "-p", "env=prod", "-m", "testdata/01_testmodule", "init", "-no-color"},
					want: []string{"Terraform has been successfully initialized!"},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config-prod.yaml", "-p", "env=prod", "-m", "testdata/01_testmodule", "plan", "-no-color"},
					want: []string{
						"# null_resource.echo will be created",
						"Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config-prod.yaml", "-p", "env=prod", "-m", "testdata/01_testmodule", "apply", "-auto-approve", "-no-color"},
					want: []string{`foo = 42
mapvar = {
  entry1 = {
    value1 = testvalue1
    value2 = true
  }
  entry2 = {
    value1 = testvalue2
    value2 = false
  }
}`},
				},
			},
		},
		{
			name: "backend check",
			runs: []testRun{
				{
					args: []string{"-d", "-c", "testdata/test-config-prod.yaml", "-p", "env=prod", "-m", "testdata/01_testmodule", "init", "-no-color"},
					want: []string{"Terraform has been successfully initialized!"},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config-prod.yaml", "-p", "env=prod", "-m", "testdata/01_testmodule", "plan", "-no-color"},
					want: []string{
						"# null_resource.echo will be created",
						"Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config-dev.yaml", "-p", "env=dev", "-m", "testdata/01_testmodule", "plan", "-no-color"},
					want: []string{
						"Configured backend does not match current environment",
						"Got: .terraform/terraform-testmodule-prod.tfstate",
						"Want: .terraform/terraform-testmodule-dev.tfstate",
						"Run terraform init -reconfigure!",
					},
					wantErr: true,
				},
				{
					args: []string{"-d", "-c", "testdata/test-config-dev.yaml", "-p", "env=dev", "-m", "testdata/01_testmodule", "init", "-reconfigure"},
					want: []string{
						"Terraform has been successfully initialized!",
					},
				},
				{
					args: []string{"-d", "-c", "testdata/test-config-dev.yaml", "-p", "env=dev", "-m", "testdata/01_testmodule", "plan", "-no-color"},
					want: []string{
						"# null_resource.echo will be created",
						"Plan: 1 to add, 0 to change, 0 to destroy.",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.RemoveAll("testdata/01_testmodule/.terraform")

			tempDir, err := ioutil.TempDir("", "gotf")
			panicOnError(err)
			defer os.RemoveAll(tempDir)

			panicOnError(os.Setenv("XDG_CACHE_HOME", filepath.Join(tempDir, "test")))

			for _, run := range tt.runs {
				got, err := runGotf(run.args)
				if run.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
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
