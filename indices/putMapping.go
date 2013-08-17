// Copyright 2013 Matthew Baird
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package indices

import (
	"encoding/json"
	"fmt"
	api "github.com/sourcegraph/elastigo/api"
	"reflect"
	"strings"
)

type MappingOptions struct {
	Timestamp  TimestampOptions             `json:"_timestamp"`
	Properties map[string]map[string]string `json:"properties"`
}

type TimestampOptions struct {
	Enabled bool `json:"enabled"`
}

func PutMapping(index string, typeName string, instance interface{}, opt MappingOptions) error {
	if opt.Properties == nil {
		opt.Properties = make(map[string]map[string]string)
	}

	instanceType := reflect.TypeOf(instance)
	if instanceType.Kind() != reflect.Struct {
		return fmt.Errorf("instance kind was not struct")
	}

	n := instanceType.NumField()
	for i := 0; i < n; i++ {
		field := instanceType.Field(i)

		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			continue
		} else if name == "" {
			name = field.Name
		}

		attrMap := make(map[string]string)
		tag := field.Tag.Get("elastic")
		if tag == "" {
			continue
		}
		attrs := strings.Split(tag, ",")
		for _, attr := range attrs {
			keyvalue := strings.Split(attr, ":")
			attrMap[keyvalue[0]] = keyvalue[1]
		}
		opt.Properties[name] = attrMap
	}

	body, err := json.Marshal(map[string]MappingOptions{typeName: opt})
	if err != nil {
		return err
	}

	_, err = api.DoCommand("PUT", fmt.Sprintf("/%s/%s/_mapping", index, typeName), string(body))
	if err != nil {
		return err
	}

	return nil
}
