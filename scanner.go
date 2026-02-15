package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Result struct {
	Refs    map[string][]string `json:"code_refs"`
	Missing map[string][]string `json:"missing"`
	Drift   []Drift             `json:"drift,omitempty"`
}

type Drift struct {
	Var   string   `json:"var"`
	In    []string `json:"in"`
	NotIn []string `json:"not_in"`
}

var pats = []*regexp.Regexp{
	regexp.MustCompile(`(?:os\.getenv|os\.environ\.get|os\.Getenv|System\.getenv)\(["'](\w+)["']\)`),
	regexp.MustCompile(`(?:os\.environ|process\.env|ENV)\[["'](\w+)["']\]`),
	regexp.MustCompile(`process\.env\.([A-Z_][A-Z0-9_]*)`),
	regexp.MustCompile(`ENV\.fetch\(["'](\w+)["']\)`),
}

var exts = map[string]bool{".py": true, ".js": true, ".ts": true, ".go": true, ".rb": true, ".java": true}

func ScanCode(dir string) map[string][]string {
	refs := map[string][]string{}
	filepath.Walk(dir, func(p string, i os.FileInfo, e error) error {
		if e != nil || i.IsDir() || !exts[filepath.Ext(p)] {
			return nil
		}
		data, _ := os.ReadFile(p)
		for _, re := range pats {
			for _, m := range re.FindAllStringSubmatch(string(data), -1) {
				if !has(refs[m[1]], p) {
					refs[m[1]] = append(refs[m[1]], p)
				}
			}
		}
		return nil
	})
	return refs
}

func ParseDotEnv(path string) map[string]string {
	m := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		return m
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l == "" || l[0] == '#' {
			continue
		}
		if k, v, ok := strings.Cut(l, "="); ok {
			m[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return m
}

func ParseDockerCompose(path string) map[string]string {
	m := map[string]string{}
	data, err := os.ReadFile(path)
	if err != nil {
		return m
	}
	var doc struct {
		Services map[string]struct {
			Env interface{} `yaml:"environment"`
		} `yaml:"services"`
	}
	if yaml.Unmarshal(data, &doc) != nil {
		return m
	}
	for _, svc := range doc.Services {
		switch env := svc.Env.(type) {
		case map[string]interface{}:
			for k, v := range env {
				m[k] = fmt.Sprintf("%v", v)
			}
		case []interface{}:
			for _, item := range env {
				k, v, _ := strings.Cut(fmt.Sprintf("%v", item), "=")
				m[k] = v
			}
		}
	}
	return m
}

func ParseK8sConfigMap(path string) map[string]string {
	var doc struct{ Data map[string]string `yaml:"data"` }
	if d, err := os.ReadFile(path); err == nil {
		yaml.Unmarshal(d, &doc)
	}
	if doc.Data == nil {
		return map[string]string{}
	}
	return doc.Data
}

func Compare(refs map[string][]string, srcs map[string]map[string]string) Result {
	r := Result{Refs: refs, Missing: map[string][]string{}}
	for v := range refs {
		for n, s := range srcs {
			if _, ok := s[v]; !ok {
				r.Missing[v] = append(r.Missing[v], n)
			}
		}
	}
	all := map[string]bool{}
	for _, s := range srcs {
		for k := range s {
			all[k] = true
		}
	}
	for v := range all {
		var in, out []string
		for n, s := range srcs {
			if _, ok := s[v]; ok {
				in = append(in, n)
			} else {
				out = append(out, n)
			}
		}
		if len(out) > 0 {
			r.Drift = append(r.Drift, Drift{v, in, out})
		}
	}
	return r
}

func has(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
