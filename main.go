package main

import (
	"log"
    "fmt"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "path/filepath"
    "os"
    "sort"
    "strings"
    "strconv"
    "github.com/parnurzeal/gorequest"
    "regexp"
    "net/http"
    "encoding/json"
    "github.com/gorilla/mux"

    "io"
)

// Config that represent the yml structure
type Config struct {
    URL string
    Matcher string
}

func parseYaml(filename string) Config {
    var config Config
    source, err := ioutil.ReadFile(filename)
    if err != nil {
        panic(err)
    }
    err = yaml.Unmarshal(source, &config)
    if err != nil {
        panic(err)
    }
    return config
}

func listYamlFiles(dir string) []string {
	var files []string

    err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
        	files = append(files, path)
    	}
        return nil
    })
    if err != nil {
        panic(err)
    }
    return files
}

func parseVersion(s string, width int) int64 {
	strList := strings.Split(s, ".")
	format := fmt.Sprintf("%%s%%0%ds", width)
	v := ""
	for _, value := range strList {
		v = fmt.Sprintf(format, v, value)
	}
	var result int64
	var err error
	if result, err = strconv.ParseInt(v, 10, 64); err != nil {
		fmt.Printf("ugh: parseVersion(%s): error=%s", s, err);
		return 0
	}
	return result;
}

// GetVersion retrieve the version of the softwar
func GetVersion(w http.ResponseWriter, r *http.Request) {
	
	var softwareVersion string
	type data struct {
		Version string
	}

	if r.Body == nil {
		fmt.Println("EMPTY")
	}

	decoder := json.NewDecoder(r.Body)

	var t data
	err := decoder.Decode(&t)

	if err == io.EOF {
		softwareVersion = ""
	} else if err != nil {
		panic(err)
	}

	if t.Version != "" {
		softwareVersion = t.Version
	}

	params := mux.Vars(r)
	softwareName := params["software"]
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	for _, yamlFile := range listYamlFiles(".") {
		if strings.Split(yamlFile, ".")[0] == softwareName {
			c := parseYaml(yamlFile)
			_, body, _ := gorequest.New().Get(c.URL).End()
			
			array := strings.Split(body, "\n")

			var availableVersions []string
			var av []string

			for _, version := range array {
				av = append(av, version)
			}

			sort.Strings(av)

			for _, v := range av {
				if strings.Contains(v, c.Matcher) {
					availableVersions = append(availableVersions, v)
				}
			}

			latestVersion := availableVersions[len(availableVersions)-1]
			r, _ := regexp.Compile("(\\d+)(?:\\.(\\d+))*")
			if softwareVersion == "" {
				if err := json.NewEncoder(w).Encode(r.FindString(latestVersion)); err != nil {
					panic(err)
				}
			} else {

				latestVersionInt64 := parseVersion(r.FindString(latestVersion), 4)
				softwareVersionInt64 := parseVersion(softwareVersion, 4)

				if latestVersionInt64 > softwareVersionInt64 {
					w.WriteHeader(http.StatusOK)
					if err := json.NewEncoder(w).Encode(r.FindString(latestVersion)); err != nil {
						panic(err)
					}
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}
		}
	}
}


func main() {
	router := mux.NewRouter()
	router.HandleFunc("/{software}", GetVersion).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}