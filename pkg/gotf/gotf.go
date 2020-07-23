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
)

var (
	Version   = "dev"
	GitCommit = "HEAD"
	BuildDate = "unknown"

	hashicorpPGPKey = []byte(`-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

mQENBFMORM0BCADBRyKO1MhCirazOSVwcfTr1xUxjPvfxD3hjUwHtjsOy/bT6p9f
W2mRPfwnq2JB5As+paL3UGDsSRDnK9KAxQb0NNF4+eVhr/EJ18s3wwXXDMjpIifq
fIm2WyH3G+aRLTLPIpscUNKDyxFOUbsmgXAmJ46Re1fn8uKxKRHbfa39aeuEYWFA
3drdL1WoUngvED7f+RnKBK2G6ZEpO+LDovQk19xGjiMTtPJrjMjZJ3QXqPvx5wca
KSZLr4lMTuoTI/ZXyZy5bD4tShiZz6KcyX27cD70q2iRcEZ0poLKHyEIDAi3TM5k
SwbbWBFd5RNPOR0qzrb/0p9ksKK48IIfH2FvABEBAAG0K0hhc2hpQ29ycCBTZWN1
cml0eSA8c2VjdXJpdHlAaGFzaGljb3JwLmNvbT6JATgEEwECACIFAlMORM0CGwMG
CwkIBwMCBhUIAgkKCwQWAgMBAh4BAheAAAoJEFGFLYc0j/xMyWIIAIPhcVqiQ59n
Jc07gjUX0SWBJAxEG1lKxfzS4Xp+57h2xxTpdotGQ1fZwsihaIqow337YHQI3q0i
SqV534Ms+j/tU7X8sq11xFJIeEVG8PASRCwmryUwghFKPlHETQ8jJ+Y8+1asRydi
psP3B/5Mjhqv/uOK+Vy3zAyIpyDOMtIpOVfjSpCplVRdtSTFWBu9Em7j5I2HMn1w
sJZnJgXKpybpibGiiTtmnFLOwibmprSu04rsnP4ncdC2XRD4wIjoyA+4PKgX3sCO
klEzKryWYBmLkJOMDdo52LttP3279s7XrkLEE7ia0fXa2c12EQ0f0DQ1tGUvyVEW
WmJVccm5bq25AQ0EUw5EzQEIANaPUY04/g7AmYkOMjaCZ6iTp9hB5Rsj/4ee/ln9
wArzRO9+3eejLWh53FoN1rO+su7tiXJA5YAzVy6tuolrqjM8DBztPxdLBbEi4V+j
2tK0dATdBQBHEh3OJApO2UBtcjaZBT31zrG9K55D+CrcgIVEHAKY8Cb4kLBkb5wM
skn+DrASKU0BNIV1qRsxfiUdQHZfSqtp004nrql1lbFMLFEuiY8FZrkkQ9qduixo
mTT6f34/oiY+Jam3zCK7RDN/OjuWheIPGj/Qbx9JuNiwgX6yRj7OE1tjUx6d8g9y
0H1fmLJbb3WZZbuuGFnK6qrE3bGeY8+AWaJAZ37wpWh1p0cAEQEAAYkBHwQYAQIA
CQUCUw5EzQIbDAAKCRBRhS2HNI/8TJntCAClU7TOO/X053eKF1jqNW4A1qpxctVc
z8eTcY8Om5O4f6a/rfxfNFKn9Qyja/OG1xWNobETy7MiMXYjaa8uUx5iFy6kMVaP
0BXJ59NLZjMARGw6lVTYDTIvzqqqwLxgliSDfSnqUhubGwvykANPO+93BBx89MRG
unNoYGXtPlhNFrAsB1VR8+EyKLv2HQtGCPSFBhrjuzH3gxGibNDDdFQLxxuJWepJ
EK1UbTS4ms0NgZ2Uknqn1WRU1Ki7rE4sTy68iZtWpKQXZEJa0IGnuI2sSINGcXCJ
oEIgXTMyCILo34Fa/C6VCm2WBgz9zZO8/rHIiQm1J5zqz0DrDwKBUM9C
=LYpS
-----END PGP PUBLIC KEY BLOCK-----`)

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
				installer := terraform.NewInstaller(urlTemplates, cfg.TerraformVersion, hashicorpPGPKey, cacheDir)
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
	tf := terraform.NewTerraform(cfg, args.ModuleDir, args.Params, args.SkipBackendCheck, shell, tfBinary)
	return tf.Execute(args.Args...)
}
