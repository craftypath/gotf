# gotf

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) ![CI](https://github.com/craftypath/gotf/workflows/CI/badge.svg?branch=master&event=push)

`gotf` is a Terraform wrapper that makes it easy to support multiple configurations, e. g. for different environments.

## Installation

### GitHub Release

Download a release from GitHub:

https://github.com/craftypath/gotf/releases

### Homebrew

```console
$ brew tap craftypath/tap
$ brew install gotf
```

## Usage

```console
$ gotf --help

  ___   __  ____  ____
 / __) /  \(_  _)(  __)
( (_ \(  O ) )(   ) _)
 \___/ \__/ (__) (__)   v0.6.0 (commit=c2206ae3c8fb02ddf32ce49c267e8d92624c37f1, date=2020-01-29T22:13:46Z)

gotf is a Terraform wrapper facilitating configurations for various environments

Usage:
  gotf [flags] [Terraform args]

Flags:
  -c, --config string       Config file to be used (default "gotf.yaml")
  -d, --debug               Print additional debug output to stderr
  -h, --help                help for gotf
  -m, --module-dir string   The module directory to run Terraform in
  -p, --params key=value    Params for templating in the config file. May be specified multiple times (default map[])
      --version             version for gotf
```

### Configuration

`gotf` is configured via config file.
By default, `gotf.yaml` is loaded from the current directory.
In a real-world scenario, you will probably have a config file per environment.
Config files support templating as specified below.

#### Parameters

##### terraformVersion

Optionally sets a specific Terraform version to use.
`gotf` will download this version and cache it in `$XDG_CACHE_HOME/gotf/terraform/<version>` verifying GPG signature and SHA256 sum.

##### varFiles

Variable files for Terraform which are added to the Terraform environment via `TF_CLI_ARGS_<command>=-var-file=<file>` for commands that support them.
They are resolved relative to this config file.

##### vars

A list of variables that are added to the Terraform environment via `TF_VAR_<var>=value` for commands that support them.

##### envs

Environment variables to be added to the Terraform process.

##### backendConfigs

Backend configs are always added as variables (`TF_VAR_backend_<var>=value`) for commands that support them and, in case of `init`, additionally as `-backend-config` CLI options.

#### Example

```yaml
terraformVersion: 0.12.21

varFiles:
  - test-{{ .Params.env }}.tfvars

vars:
  foo: foovalue
  templated_var: "{{ .Params.param }}"
  mapvar: |-
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
  module_path: "{{ .Params.moduleDir }}"
  module: "{{ base .Params.moduleDir }}"
  state_key_prefix: '{{ (splitn "_" 2 (base .Params.moduleDir))._1 }}'

envs:
  BAR: barvalue
  TEMPLATED_ENV: "{{ .Params.param }}"

backendConfigs:
  key: "{{ .Vars.state_key_prefix }}_{{ .Vars.templated_var }}_{{ .Params.key_suffix }}"
  storage_account_name: be_storage_account_name_{{ .Vars.foo }}_{{ .Envs.BAR }}
  resource_group_name: be_resource_group_name_{{ .Vars.foo }}_{{ .Envs.BAR }}
  container_name: be_container_name_{{ .Vars.foo }}_{{ .Envs.BAR }}
```

#### Templating

Go templating can be used in the config file as follows.
The [Sprig](https://masterminds.github.io/sprig/) function library is included.

* In the first templating pass, `varFiles`, `vars`, and `envs` are processed.
  All parameters specified using the `-p|--param` flag are available in the `.Params` object.
  The module directory passed with the `--module-dir|-m` parameter is available as `module` dir in the `.Params` object.
* In the second templating pass, `backendConfigs` are processed.
  `vars` are available as `.Vars`, `envs` are available as `.Envs` with the results from the first templating pass.
  Additionally, `.Params` is also available again.

Using the above config file, running `terraform init` could look like this:

```console
$ gotf -c example-config.yaml -p param=myval -p key_suffix=mysuffix -m my_modules/01_testmodule init
```

After processing, the config file would look like this:

```yaml
terraformVersion: 0.12.21

varFiles:
  - test-prod.tfvars

vars:
  foo: foovalue
  templated_var: "myval"
  mapvar: |-
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
  module_path: "my_modules/01_testmodule"
  module: "01_testmodule"
  state_key_prefix: 'testmodule'

envs:
  BAR: barvalue
  TEMPLATED_ENV: "myval"

backendConfigs:
  key: testmodule_myval_mysuffix
  storage_account_name: be_storage_account_name_foovalue_barvalue
  resource_group_name: be_resource_group_name_foovalue_barvalue
  container_name: be_container_name_foovalue_barvalue
```

#### Debug Output

Specifying the `--debug` flag produces debug output which is written to stderr.
For example, the integration test in [cmd/gotf/gotf_test.go](cmd/gotf/gotf_test.go) produces the following debug output before running Terraform:

```console
gotf> Loading config file: testdata/test-config-prod.yaml
gotf> Processing varFiles...
gotf> Processing vars...
gotf> Processing envs...
gotf> Proessing backendConfigs...
gotf> Using Terraform version 0.12.21
gotf> Downloading Terraform distro...
gotf> Downloading SHA256 sums file...
gotf> Downloading SHA256 sums signature file...
gotf> Verifying GPG signature...
gotf> Verifying SHA256 sum...
gotf> Unzipping distro...
gotf> Terraform binary: /Users/myuser/Library/Caches/gotf/terraform/0.12.21/terraform
gotf>
gotf> Terraform command-line:
gotf> -----------------------
gotf> /Users/myuser/Library/Caches/gotf/terraform/0.12.21/terraform init -no-color
gotf>
gotf> Terraform environment:
gotf> ----------------------
gotf> TF_VAR_mapvar={
  entry1 = {
    value1 = testvalue1
    value2 = true
  }
  entry2 = {
    value1 = testvalue2
    value2 = false
  }
}
gotf> TF_VAR_backend_path=.terraform/terraform-testmodule-prod.tfstate
gotf> TF_CLI_ARGS_init=-backend-config=path=".terraform/terraform-testmodule-prod.tfstate"
gotf> TF_VAR_state_key=testmodule
gotf> TF_VAR_foo=42
gotf> BAR=barvalue
gotf> TF_CLI_ARGS_apply=-var-file="../01_testmodule/test-prod.tfvars"
gotf> TF_CLI_ARGS_destroy=-var-file="../01_testmodule/test-prod.tfvars"
gotf> TF_CLI_ARGS_plan=-var-file="../01_testmodule/test-prod.tfvars"
gotf> TF_CLI_ARGS_refresh=-var-file="../01_testmodule/test-prod.tfvars"
gotf> TF_CLI_ARGS_import=-var-file="../01_testmodule/test-prod.tfvars"
gotf> TF_VAR_module_dir=01_testmodule
```
