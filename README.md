# gotf

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) ![CI](https://github.com/craftypath/gotf/workflows/CI/badge.svg?branch=master&event=push)

`gotf` is a Terraform wrapper that makes it easy to support multiple configurations, e.g. for different environments.

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
 \___/ \__/ (__) (__)   v0.13.0 (commit=77e3ad0, date=2021-05-07T22:03:44Z)

gotf is a Terraform wrapper facilitating configurations for various environments

Usage:
  gotf [flags] [Terraform args]

Flags:
  -c, --config string        Config file to be used (default "gotf.yaml")
  -d, --debug                Print additional debug output to stderr
  -h, --help                 help for gotf
  -m, --module-dir string    The module directory to run Terraform in
  -n, --no-vars              Don't add any variables when running Terraform.
                             This is necessary when running 'terraform apply' with a plan file.
  -p, --params key=value     Params for templating in the config file. May be specified multiple times (default map[])
  -s, --skip-backend-check   Skip checking for changed backend configuration
  -v, --version              version for gotf
```

## Demo

Check out the [demo](demo) project which does not use cloud providers and keeps state locally.
This allows you to play with the tool on your local machine.

## Talks

<a href="https://www.youtube.com/watch?v=Tel-9m5_MSc&list=PLD9VybHH2wnYtuaiEcjMH5n2gTgNSDYAV" target="_blank"><img src="https://i.ytimg.com/vi/Tel-9m5_MSc/maxresdefault.jpg"
alt="🇩🇪 gotf – Umgebungen mit Terraform einfacher managen | Reinhard von codecentric auf der #SoftwerkerKonf" width="240" height="135" border="10" /></a><br>
**🇩🇪 gotf – Umgebungen mit Terraform einfacher managen | Reinhard von codecentric auf der #SoftwerkerKonf**

## Configuration

`gotf` is configured via config file.
By default, `gotf.yaml` is loaded from the current directory.
Config files support templating as specified below.

### Parameters

#### `terraformVersion`

Optionally sets a specific Terraform version to use.
`gotf` will download this version and cache it in `$XDG_CACHE_HOME/gotf/terraform/<version>` verifying GPG signature and SHA256 sum.

#### `params`

Config entries that can be used for templating. See section on templating below for details.

#### `requiredParams`

In addition to specifying `params` in the config file, they may also be specified on the command-line using the `-p|--param` flag.
Params that are required can be configured here.
Allowed values for a `param` must be specified as list.
If no restrictions apply, no value or an empty list must be specified.
Values must be strings.

#### `globalVarFiles`

A list of variables files which are added to the Terraform environment via `TF_CLI_ARGS_<command>=-var-file=<file>` for commands that support them.
They are resolved relative to this config file.

#### `moduleVarFiles.<moduleDir>`

A list of module-specific variables files which are added to the Terraform environment if the corresponding module is run via `TF_CLI_ARGS_<command>=-var-file=<file>` for commands that support them.
They are resolved relative to this config file.

#### `globalVars`

Variables which are added to the Terraform environment via `TF_VAR_<var>=value` for commands that support them.

#### `moduleVars.<moduleDir>`

Module-specific variables which are added to the Terraform environment if the corresponding module is run via `TF_VAR_<var>=value` for commands that support them.
Module-specific variables override global ones.

#### `envs`

Environment variables to be added to the Terraform process.

#### `backendConfigs`

Backend configuration added as `-backend-config` CLI options when the Terraform `init` command is run.

#### `ignoreMissingVarFiles`

If set to `true`, gotf checks whether configured variable files exist and does not pass them to Terraform if they don't.

### Example

```yaml
terraformVersion:  1.0.9

requiredParams:
  environment:
    - dev
    - prod

params:
  param: myval

globalVarFiles:
  - global-{{ .Params.environment }}.tfvars
  - global.tfvars

globalVars:
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
  module_dir: "{{ .Params.moduleDir }}"
  state_key: '{{ (splitn "_" 2 .Params.moduleDir)._1 }}'

moduleVarFiles:
  01_networking:
    - 01_networking/{{ .Params.environment }}.tfvars
  02_compute:
    - 02_compute/{{ .Params.environment }}.tfvars

moduleVars:
  01_networking:
    myvar: value for networking
  02_compute:
    myvar: value for compute

envs:
  BAR: barvalue
  TEMPLATED_ENV: "{{ .Params.param }}"

backendConfigs:
  key: "{{ .Vars.state_key }}"
  storage_account_name: mytfstateaccount{{ .Params.environment }}
  resource_group_name: mytfstate-{{ .Params.environment }}
  container_name: mytfstate-{{ .Params.environment }}
```

### Templating

Go templating can be used in the config file as follows.
Hermetic, i.e. repeatable, functions from the [Sprig](https://masterminds.github.io/sprig/) function library are included.

* In the first templating pass, `globalVarFiles`, `globalVars`, `moduleVarFiles`, `moduleVars`, and `envs` are processed.
  All parameters specified under `params` and using the `-p|--param` flag are available in the `.Params` object.
  CLI params override those specified in the config file.
  The basename of the module directory passed with the `--module-dir|-m` parameter is available as `moduleDir` dir in the `.Params` object.
* In the second templating pass, `backendConfigs` are processed.
  `globalVars` and ` moduleVars` are available as `.Vars` and `envs` are available as `.Envs` with the results from the first templating pass.
  Additionally, `.Params` is also available again.

Using the above config file, running `terraform init` could look like this:

```console
$ gotf -c example-config.yaml -p environment=dev -m 01_networking init
```

After processing, the config file would look like this:

```yaml
terraformVersion:  1.0.9

requiredParams:
  environment:
    - dev
    - prod

params:
  param: myval

globalVarFiles:
  - global-dev.tfvars
  - global.tfvars

globalVars:
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
  module_dir: "01_networking"
  state_key: 'networking'

moduleVarFiles:
  01_networking:
    - 01_networking/dev.tfvars
  02_compute:
    - 02_compute/dev.tfvars

moduleVars:
  01_networking:
    myvar: value for networking
  02_compute:
    myvar: value for compute

envs:
  BAR: barvalue
  TEMPLATED_ENV: "myval"

backendConfigs:
  key: "networking"
  storage_account_name: mytfstateaccountdev
  resource_group_name: mytfstate-dev
  container_name: mytfstate-dev
```

## Debug Output

Specifying the `--debug` flag produces debug output which is written to stderr.
For example, the integration test in [cmd/gotf/gotf_test.go](cmd/gotf/gotf_test.go) produces the following debug output before running Terraform:

```console
gotf> Loading config file: testdata/test-config.yaml
gotf> Processing var files...
gotf> Processing global var files...
gotf> Processing module var files...
gotf> Processing global vars...
gotf> Processing module vars...
gotf> Processing envs...
gotf> Processing backend configs...
gotf> Using Terraform version 1.0.9
gotf> Terraform version 1.0.9 already installed.
gotf> Terraform binary: /Users/myuser/Library/Caches/gotf/terraform/1.0.9/terraform
gotf>
gotf> Terraform command-line:
gotf> -----------------------
gotf> /Users/myuser/Library/Caches/gotf/terraform/1.0.9/terraform init -no-color
gotf>
gotf> Terraform environment:
gotf> ----------------------
gotf> TF_CLI_ARGS_destroy=-var-file="../global-prod.tfvars" -var-file="../global.tfvars" -var-file="prod.tfvars"
gotf> TF_VAR_templated_var=myval
gotf> BAR=barvalue
gotf> TF_CLI_ARGS_apply=-var-file="../global-prod.tfvars" -var-file="../global.tfvars" -var-file="prod.tfvars"
gotf> TF_VAR_state_key=networking
gotf> TF_VAR_myvar=value for networking
gotf> TF_CLI_ARGS_plan=-var-file="../global-prod.tfvars" -var-file="../global.tfvars" -var-file="prod.tfvars"
gotf> TF_CLI_ARGS_import=-var-file="../global-prod.tfvars" -var-file="../global.tfvars" -var-file="prod.tfvars"
gotf> TF_VAR_foo=42
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
gotf> TF_VAR_module_dir=01_networking
gotf> TF_CLI_ARGS_init=-backend-config=path=".terraform/terraform-networking-prod.tfstate"
gotf> TEMPLATED_ENV=myval
gotf> TF_CLI_ARGS_refresh=-var-file="../global-prod.tfvars" -var-file="../global.tfvars" -var-file="prod.tfvars"
```
