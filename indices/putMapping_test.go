package indices

import (
	"encoding/json"
	api "github.com/sourcegraph/elastigo/api"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
)

func setup(t *testing.T) {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	api.Domain = strings.Split(serverURL.Host, ":")[0]
	api.Port = strings.Split(serverURL.Host, ":")[1]
}

func teardown() {
	server.Close()
}

type TestStruct struct {
	Id            string `json:"id" elastic:"index:not_analyzed"`
	DontIndex     string `json:"dontIndex" elastic:"index:no"`
	Number        int    `json:"number" elastic:"type:integer,index:analyzed"`
	Omitted       string `json:"-"`
	NoJson        string `elastic:"type:string"`
	unexported    string
	JsonOmitEmpty string `json:"jsonOmitEmpty,omitempty" elastic:"type:string"`
	Embedded
	MultiAnalyze string `json:"multi_analyze"`
}

type Embedded struct {
	EmbeddedField string `json:"embeddedField" elastic:"type:string"`
}

func TestPutMapping(t *testing.T) {
	setup(t)
	defer teardown()

	options := MappingOptions{
		Timestamp: TimestampOptions{Enabled: true},
		Id:        IdOptions{Index: "analyzed", Path: "id"},
		Properties: map[string]interface{}{
			// special properties that can't be expressed as tags
			"multi_analyze": map[string]interface{}{
				"type": "multi_field",
				"fields": map[string]map[string]string{
					"ma_analyzed":   {"type": "string", "index": "analyzed"},
					"ma_unanalyzed": {"type": "string", "index": "un_analyzed"},
				},
			},
		},
	}
	expValue := MappingForType("myType", MappingOptions{
		Timestamp: TimestampOptions{Enabled: true},
		Id:        IdOptions{Index: "analyzed", Path: "id"},
		Properties: map[string]interface{}{
			"NoJson":        map[string]string{"type": "string"},
			"dontIndex":     map[string]string{"index": "no"},
			"embeddedField": map[string]string{"type": "string"},
			"id":            map[string]string{"index": "not_analyzed"},
			"jsonOmitEmpty": map[string]string{"type": "string"},
			"number":        map[string]string{"index": "analyzed", "type": "integer"},
			"multi_analyze": map[string]interface{}{
				"type": "multi_field",
				"fields": map[string]map[string]string{
					"ma_analyzed":   {"type": "string", "index": "analyzed"},
					"ma_unanalyzed": {"type": "string", "index": "un_analyzed"},
				},
			},
		},
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var value Mapping
		json.NewDecoder(r.Body).Decode(&value)

		expValJson, _ := json.MarshalIndent(expValue, "", "  ")
		valJson, _ := json.MarshalIndent(value, "", "  ")

		if string(expValJson) != string(valJson) {
			t.Errorf("Expected %s but got %s", string(expValJson), string(valJson))
		}
	})

	err := PutMapping("myIndex", "myType", TestStruct{}, options)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
