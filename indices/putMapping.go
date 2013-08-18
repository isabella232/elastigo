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

type Mapping map[string]MappingOptions

type MappingOptions struct {
	Timestamp  TimestampOptions             `json:"_timestamp"`
	Id         IdOptions                    `json:"_id"`
	Properties map[string]map[string]string `json:"properties"`
}

type TimestampOptions struct {
	Enabled bool `json:"enabled"`
}

type IdOptions struct {
	Index string `'json:"index"`
	Path  string `json:"path"`
}

func (m_ Mapping) Options() MappingOptions {
	m := map[string]MappingOptions(m_)
	for _, v := range m {
		return v
	}
	panic(fmt.Errorf("Malformed input: %v", m_))
}

func MappingForType(typeName string, opts MappingOptions) Mapping {
	return map[string]MappingOptions{typeName: opts}
}

func PutMapping(index string, typeName string, instance interface{}, opt MappingOptions) error {
	instanceType := reflect.TypeOf(instance)
	if instanceType.Kind() != reflect.Struct {
		return fmt.Errorf("instance kind was not struct")
	}

	opt.Properties = make(map[string]map[string]string)
	GetProperties(instanceType, opt.Properties)

	body, err := json.Marshal(MappingForType(typeName, opt))
	if err != nil {
		return err
	}

	_, err = api.DoCommand("PUT", fmt.Sprintf("/%s/%s/_mapping", index, typeName), string(body))
	if err != nil {
		return err
	}

	return nil
}

func GetProperties(t reflect.Type, prop map[string]map[string]string) {
	n := t.NumField()
	for i := 0; i < n; i++ {
		field := t.Field(i)

		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			continue
		} else if name == "" {
			name = field.Name
		}

		attrMap := make(map[string]string)
		tag := field.Tag.Get("elastic")
		if tag == "" {
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				GetProperties(field.Type, prop)
			}
			continue
		}
		attrs := strings.Split(tag, ",")
		for _, attr := range attrs {
			keyvalue := strings.Split(attr, ":")
			attrMap[keyvalue[0]] = keyvalue[1]
		}
		prop[name] = attrMap
	}
}
