# gotf

`gotf` is a Terraform wrapper that makes it easy to support multiple configurations, e. g. for different environments.

## Installation

### GitHub Release

Download a release from GitHub:

https://github.com/unguiculus/gotf/releases

### Using `go get`

```console
$ go get github.com/unguiculus/gotf
```

## Usage

```console
$ gotf --help

  ___   __  ____  ____
 / __) /  \(_  _)(  __)
( (_ \(  O ) )(   ) _)
 \___/ \__/ (__) (__)   v0.2.0 (commit=eb8844da0920b4ef912c6f719d1b77495452e94f, date=2020-01-24T23:14:54Z)

gotf is a Terraform wrapper facilitating configurations for various environments

Usage:
  gotf [flags] [Terraform args]

Flags:
  -c, --config string      Config file to be used (default "gotf.yaml")
  -d, --debug              Print additional debug output
  -h, --help               help for gotf
  -p, --params key=value   Params for templating in the config file. May be specified multiple times (default map[])
      --version            version for gotf
```

`gotf` is configured via config file.
By default, `gotf.yaml` is loaded from the current directory.

### Example

```yaml
# Optionally set a specific Terraform version. gotf will download this version and cache
# it in $XDG_CACHE_HOME/gotf/terraform/<version> verifying GPG signature and SHA256 sum
terraformVersion: 0.12.20

# tfvars files are added to the Terraform environment via
# TF_CLI_ARGS_<command>=-var-file=<file> for commands that support them
varFiles:
  - test-{{ .Params.env }}.tfvars

# Variables are added to the Terraform environment via
# TF_VAR_<var>=value for commands that support them
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
  state_key_prefix: '{{ regexSplit "\\d+_" (base .Params.moduleDir) 2 | last }}'

# Environment variables are added to the Terraform calls environment
envs:
  BAR: barvalue
  TEMPLATED_ENV: "{{ .Params.param }}"

# Backend configs are always added as variables (TF_VAR_backend_<var>=value) for commands
# that support them and, if in case of 'init' additionally as '-backend-config' CLI options.
# Note the prefix 'backend_'  in the variable names.
backendConfigs:
  key: "{{ .Vars.state_key_prefix }}_{{ .Vars.templated_var }}_{{ .Params.key_suffix }}
  storage_account_name: be_storage_account_name_{{ .Vars.foo }}_{{ .Envs.BAR }}
  resource_group_name: be_resource_group_name_{{ .Vars.foo }}_{{ .Envs.BAR }}
  container_name: be_container_name_{{ .Vars.foo }}_{{ .Envs.BAR }}
```

## Templating

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
terraformVersion: 0.12.20

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
