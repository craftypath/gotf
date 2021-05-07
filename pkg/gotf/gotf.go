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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/craftypath/gotf/pkg/config"
	"github.com/craftypath/gotf/pkg/sh"
	terraform "github.com/craftypath/gotf/pkg/tf"

	_ "embed"
)

var (
	Version   = "dev"
	GitCommit = "HEAD"
	BuildDate = "unknown"

	//go:embed gpg_key_old.txt
	hashicorpPGPKeyOld []byte

	//go:embed gpg_key_new.txt
	hashicorpPGPKeyNew []byte

	urlTemplates = &terraform.URLTemplates{
		TargetFile:              "https://releases.hashicorp.com/terraform/%[1]s/terraform_%[1]s_%s_%s.zip",
		SHA256SumsFile:          "https://releases.hashicorp.com/terraform/%[1]s/terraform_%[1]s_SHA256SUMS",
		SHA256SumsSignatureFile: "https://releases.hashicorp.com/terraform/%[1]s/terraform_%[1]s_SHA256SUMS.sig",
	}
)

type Args struct {
	Debug            bool
	ConfigFile       string
	ModuleDir        string
	Params           map[string]string
	SkipBackendCheck bool
	NoVars           bool
	Args             []string
}

func Run(args Args) error {
	if len(args.Args) == 0 {
		return errors.New("no arguments for Terraform specified")
	}

	if args.Debug {
		log.SetOutput(os.Stderr)
		log.SetFlags(0)
		log.SetPrefix("gotf> ")
	} else {
		log.SetOutput(ioutil.Discard)
	}

	cfg, err := config.Load(args.ConfigFile, args.ModuleDir, args.Params)
	if err != nil {
		return fmt.Errorf("could not load config file %q: %w", args.ConfigFile, err)
	}

	var tfBinary string
	if cfg.TerraformVersion != "" {
		log.Println("Using Terraform version", cfg.TerraformVersion)
		cacheDir, err := xdg.CacheFile(filepath.Join("gotf", "terraform", cfg.TerraformVersion))
		if err != nil {
			return err
		}
		tfBinary = filepath.Join(cacheDir, "terraform")
		if _, err := os.Stat(tfBinary); err != nil {
			if os.IsNotExist(err) {
				installer := terraform.NewInstaller(urlTemplates, cfg.TerraformVersion, [][]byte{hashicorpPGPKeyNew, hashicorpPGPKeyOld}, cacheDir)
				if err = installer.Install(); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			log.Println("Terraform version", cfg.TerraformVersion, "already installed.")
		}
	} else {
		tfBinary = "terraform"
	}

	log.Println("Terraform binary:", tfBinary)

	shell := sh.Shell{}
	tf := terraform.NewTerraform(cfg, args.ModuleDir, args.Params, args.SkipBackendCheck, args.NoVars, shell, tfBinary)
	return tf.Execute(args.Args...)
}
