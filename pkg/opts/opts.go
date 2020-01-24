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

package opts

import (
	"fmt"
	"strings"
)

type MapOpts struct {
	values    map[string]string
}

func (o *MapOpts) Set(value string) error {
	pair := strings.SplitN(value, "=", 2)
	if len(pair) == 1 {
		return fmt.Errorf("no value specified for %q", pair[0])
	}
	o.values[pair[0]] = pair[1]
	return nil
}

func (o *MapOpts) GetAll() map[string]string {
	return o.values
}

func (o *MapOpts) String() string {
	return fmt.Sprintf("%v", o.values)
}

func (o *MapOpts) Type() string {
	return "key=value"
}

func NewMapOpts() *MapOpts {
	return &MapOpts{
		values:    make(map[string]string),
	}
}
