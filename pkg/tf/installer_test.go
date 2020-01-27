// Copyright The gotf Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terraform

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testGPGPublicKey = []byte(`-----BEGIN PGP PUBLIC KEY BLOCK-----

mQINBF4tYeABEACuR7HyivrLNFWZK6+gccZ7ZfTXPegpj3wMXmTPDjsvAWH2f9sq
dtE1QLBy5YFeloY3xcIIf1inccJr6sXsmA8h8JGcp8uiQFsY/gJq/rXzZuJFC7bU
+yCkZVwm0AXrf9TUWZyr1o+/EX9gVd782BwCBmVTDTV/IzU2nPRBxrwGV32gKEDX
twn58UAxySjAykkfqu7C9nnQtaiqT48RO3L28ZTDD+WGVUJBuPTq9nRtp8HNKst9
dOdJYPUXnE9dKEynXB5nWYh+G31IEgrfDXJGl9BMZbyZAGTLqbQJIR7aWa6PD6xf
sGVHY0UA3woa/i2QbJZXqTvjPcXsqP4aIoDk/nTHatYHmin+m7nM1cneyF2WZaHi
sz9JIVgUfxgcwqCGQKVAOPKJS3HE9buvCQdNAfLHAZG6oT1uXPn0zHpGhUotJjWf
K9xMPvscCrdTxyLLNSE9xNRIkrAcBLu6SArf/RJJsirDz8lvyIFyMxJuOcNmPWSH
4ZHxZaIKOQc7i4THLH4JIXecBA0oVTyl8FozZEyuYYVyFdV1X4wnHbZWXdPb1Ucb
V7rXPc0cuEuvCE20C0N//7uOtUUT5a9H6YEjy/cGqaSFRsJnngvDypnU6RLDZzYE
H+7c4AYFEPvMGT6W0iYF45xERp1Rz2TeL1NcASP40XNhonwiJCxQLwd1FwARAQAB
tCtnb3RmIHRlc3QgKGdvdGYgdGVzdCBrZXkpIDxnb3RmQGV4bXBsZS5jb20+iQJO
BBMBCAA4FiEEvHAD62leJ+eAEbV8nyHJWpRwU9AFAl4tYeACGwMFCwkIBwIGFQoJ
CAsCBBYCAwECHgECF4AACgkQnyHJWpRwU9Bo1BAAmc/AjpRXx9yw/or4dFAKjU3z
jtnTojZcww3dv3iyPqRysgtfJgm3BRZztcDNEUUq4PoKn+cbFNHJcqAghGo/nALH
DLcnesBHhmrnQtprT6g29jJSD5uA7WfqCAKKZ1jiRPm5b3adZ+HOrzjJWlK7x0Ah
0FgY7PFUtfQZgCgq/MSQv4Udw4mH1Vprwp7eDYvYwsSBLzT2CYld3Jx/PUhjbN/n
aYo0mAO2jlTKVcOoWsoiDeOWGjNmKt+sHznrvBixSFAO9cSN0EuQWkXQyJd+C/Tt
66qYVnBqqLcYRwrLYT8t6cnQj/gZcLg5IaJGO5sDKSyXWG+k+Be2NMsK+0JdLk3P
M0abAe2+5RH+yqrsWcYamikCrFF5aoNDeqxkOIx4NO/R6gzlFkcN9B226pM5ZXVe
EdJHxm7XU7SilQwCr8QWELqvxuvSX+yMpFET6730dSCAxT3KwCEQjkBY7HFDJ0Gn
8ZOYXHHP8Y3SDfM2Y4IvCpD0rpk5//Ci/WoNkztfKs2ciWQ+00gCc36xOBMKu7yE
kdbIIs9rDtIWEXh/SeiWWwyV+fzsWRMttw86f/GTrGuMiSz9bV3J6KBuf1ymd6OU
/nFJH1t7x+/Eug4u++bNhz3zjQBCO+6w+uy2s4RijlD7Cla9n5vFe/KIkjdo0922
1+PzB/G4oSzELNctXuU=
=xyGe
-----END PGP PUBLIC KEY BLOCK-----
`)

func TestInstaller_Install(t *testing.T) {
	dir, err := ioutil.TempDir("", "gotf")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	transport := &http.Transport{}
	cwd, _ := os.Getwd()
	transport.RegisterProtocol("file", http.NewFileTransport(http.Dir(cwd)))
	httpClient := &http.Client{Transport: transport}

	urlTemplates := &URLTemplates{
		TargetFile:              "file://./testdata/test_%s_%s_%s.zip",
		SHA256SumsFile:          "file://./testdata/test_%s_SHA256SUMS",
		SHA256SumsSignatureFile: "file://./testdata/test_%s_SHA256SUMS.sig",
	}
	installer := NewInstaller(urlTemplates, "0.42.0", testGPGPublicKey, dir)
	installer.httpClient = httpClient

	err = installer.Install()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(dir, "test.txt"))
}
