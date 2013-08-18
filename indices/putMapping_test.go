package indices

import (
	"encoding/json"
	api "github.com/sourcegraph/elastigo/api"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
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
	}
	expValue := map[string]MappingOptions{
		"myType": MappingOptions{
			Timestamp: TimestampOptions{Enabled: true},
			Id:        IdOptions{Index: "analyzed", Path: "id"},
			Properties: map[string]map[string]string{
				"id":            {"index": "not_analyzed"},
				"dontIndex":     {"index": "no"},
				"number":        {"type": "integer", "index": "analyzed"},
				"NoJson":        {"type": "string"},
				"jsonOmitEmpty": {"type": "string"},
				"embeddedField": {"type": "string"},
			},
		},
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var value map[string]MappingOptions
		json.NewDecoder(r.Body).Decode(&value)

		if !reflect.DeepEqual(expValue, value) {
			t.Errorf("Expected: %+v, but got: %+v", expValue, value)
		}
	})

	err := PutMapping("myIndex", "myType", TestStruct{}, options)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
