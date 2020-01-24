# gotf

`gotf` is a Terraform wrapper thst makes it easy to support multiple configuration, e. g. for different environments.

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

`goft` is configured via a config file.
By default, `gotf.yaml` is loaded from the current directory.

### Example

```yaml
varFiles:
  # tfvars files are added to the Terraform environment via
  # TF_CLI_ARGS_<command>=-var-file=<file> for commands that support them
  - testmodule/test-{{ .Params.env }}.tfvars
vars:
  # Variables are added to the Terraform environment via
  # TF_VAR_<var>=value for commands that support them
  foo: foovalue
  templatedVar: "{{ .Params.param }}"
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
envs:
  # Environment variables are added to the Terraform calls environment
  BAR: barvalue
  TEMPLATED_ENV: "{{ .Params.param }}"
backendConfigs:
  # Backend configs are always added as variables (TF_VAR_<var>=value) for commands
  # that support them and, if in case of 'init' additionally as '-backend-config' CLI options
  backend_key: be_key_{{ .Vars.foo }}_{{ .Envs.BAR }}_{{ .Vars.templatedVar }}_{{ .Params.key_suffix }}
  backend_storage_account_name: be_storage_account_name_{{ .Vars.foo }}_{{ .Envs.BAR }}
  backend_resource_group_name: be_resource_group_name_{{ .Vars.foo }}_{{ .Envs.BAR }}
  backend_container_name: be_container_name_{{ .Vars.foo }}_{{ .Envs.BAR }}
```

## Templating

Go templating can be used in the config file as follows.

* In the first templating pass, `varFiles`, `vars`, and `envs` are processed.
  All parameters specified using the `-p|--param` flag are available in the `.Params` object.
* In the second templating pass, `backendConfigs` are processed.
  `vars` are available as `.Vars`, `envs` are available as `.Envs` with the results from the first templating pass.
  Additionally, `.Params` is also available again.

Using the above config file, running `terraform init` could look like this:

```console
$ gotf -c example-config.yaml -p param=myval -p key_suffix=mysuffix init
```

After processing, the config file would look like this:

```yaml
varFiles:
  # tfvars files are added to the Terraform environment via
  # TF_CLI_ARGS_<command>=-var-file=<file> for commands that support them
  - testmodule/test-prod.tfvars
vars:
  # Variables are added to the Terraform environment via
  # TF_VAR_<var>=value for commands that support them
  foo: foovalue
  templatedVar: "myval"
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
envs:
  # Environment variables are added to the Terraform calls environment
  BAR: barvalue
  TEMPLATED_ENV: "myval"
backendConfigs:
  # Backend configs are always added as variables (TF_VAR_<var>=value) for commands
  # that support them and, if in case of 'init' additionally as '-backend-config' CLI options
  backend_key: be_key_foovalue_barvalue_myval_mysuffix
  backend_storage_account_name: be_storage_account_name_foovalue_barvalue
  backend_resource_group_name: be_resource_group_name_foovalue_barvalue
  backend_container_name: be_container_name_foovalue_barvalue
```
