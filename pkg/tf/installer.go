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

package terraform

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mholt/archiver/v3"
	"golang.org/x/crypto/openpgp"
)

type URLTemplates struct {
	TargetFile              string
	SHA256SumsFile          string
	SHA256SumsSignatureFile string
}

type Installer struct {
	urlTemplates *URLTemplates
	version      string
	gpgPublicKey []byte
	dstDir       string
	httpClient   *http.Client
}

func NewInstaller(urlTemplates *URLTemplates, version string, gpgPublicKey []byte, dstDir string) *Installer {
	return &Installer{
		urlTemplates: urlTemplates,
		version:      version,
		gpgPublicKey: gpgPublicKey,
		dstDir:       dstDir,
		httpClient:   http.DefaultClient,
	}
}

func (i *Installer) Install() error {
	if err := os.MkdirAll(i.dstDir, os.ModePerm); err != nil {
		return fmt.Errorf("could not create installation directory: %w", err)
	}

	log.Println("Downloading Terraform distro...")
	url := fmt.Sprintf(i.urlTemplates.TargetFile, i.version, runtime.GOOS, runtime.GOARCH)
	targetFilePath, err := i.download(url)
	if err != nil {
		return fmt.Errorf("could download Terraform distro: %w", err)
	}

	log.Println("Downloading SHA256 sums file...")
	url = fmt.Sprintf(i.urlTemplates.SHA256SumsFile, i.version)
	sha256sumsFilePath, err := i.download(url)
	if err != nil {
		return fmt.Errorf("could download SHA256 sums file: %w", err)
	}

	log.Println("Downloading SHA256 sums signature file...")
	url = fmt.Sprintf(i.urlTemplates.SHA256SumsSignatureFile, i.version)
	sha256sumsSignatureFilePath, err := i.download(url)
	if err != nil {
		return fmt.Errorf("could download SHA256 sums signature file: %w", err)
	}

	log.Println("Verifying GPG signature...")
	if err := i.verifyGPGSignature(sha256sumsFilePath, sha256sumsSignatureFilePath); err != nil {
		return fmt.Errorf("GPG signature verification failed: %w", err)
	}

	log.Println("Verifying SHA256 sum...")
	if err := i.verifySHA256sum(targetFilePath, sha256sumsFilePath); err != nil {
		return fmt.Errorf("SHA256 sum verification failed: %w", err)
	}

	log.Println("Unzipping distro...")
	if err := archiver.Unarchive(targetFilePath, i.dstDir); err != nil {
		return fmt.Errorf("could not unzip distro: %w", err)
	}
	return nil
}

func (i *Installer) download(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	fileName := path.Base(req.URL.Path)
	resp, err := i.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	filePath := filepath.Join(i.dstDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return filePath, err
}

func (i *Installer) verifyGPGSignature(targetFilePath string, signatureFilePath string) error {
	signature, err := os.Open(signatureFilePath)
	if err != nil {
		return err
	}

	target, err := os.Open(targetFilePath)
	if err != nil {
		return err
	}

	r := bytes.NewReader(i.gpgPublicKey)
	keyring, err := openpgp.ReadArmoredKeyRing(r)
	if err != nil {
		return err
	}

	if _, err := openpgp.CheckDetachedSignature(keyring, target, signature); err != nil {
		return err
	}

	return nil
}

func (i *Installer) verifySHA256sum(targetFilePath string, sha256sumsFilePath string) error {
	zipFileBytes, err := ioutil.ReadFile(targetFilePath)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(zipFileBytes)

	file, err := os.Open(sha256sumsFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	targetFileName := filepath.Base(targetFilePath)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, " "+targetFileName) {
			expectedSha256sum := hex.EncodeToString(hash[:])
			if strings.HasPrefix(line, expectedSha256sum+" ") {
				return nil
			}
			return errors.New("invalid sha256sum")
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return errors.New("no matching sha256sum found")
}
