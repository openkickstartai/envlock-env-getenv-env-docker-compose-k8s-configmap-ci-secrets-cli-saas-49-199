package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanCodeMultiLang(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "app.py"), []byte("import os\ndb = os.getenv('DATABASE_URL')\nkey = os.environ.get('SECRET_KEY')\n"), 0644)
	os.WriteFile(filepath.Join(dir, "server.js"), []byte("const port = process.env.PORT\nconst host = process.env['REDIS_HOST']\n"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nimport \"os\"\nvar a = os.Getenv(\"GO_VAR\")\n"), 0644)
	refs := ScanCode(dir)
	for _, want := range []string{"DATABASE_URL", "SECRET_KEY", "PORT", "REDIS_HOST", "GO_VAR"} {
		if _, ok := refs[want]; !ok {
			t.Errorf("expected %s in refs, got %v", want, refs)
		}
	}
	if len(refs) != 5 {
		t.Errorf("expected 5 vars, got %d: %v", len(refs), refs)
	}
}

func TestParseDotEnv(t *testing.T) {
	f := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(f, []byte("DATABASE_URL=postgres://localhost/db\nSECRET_KEY=abc123\n# comment line\n\nPORT=8080\n"), 0644)
	m := ParseDotEnv(f)
	if len(m) != 3 {
		t.Fatalf("expected 3 vars, got %d: %v", len(m), m)
	}
	if m["PORT"] != "8080" {
		t.Errorf("expected PORT=8080, got %q", m["PORT"])
	}
	if m["DATABASE_URL"] != "postgres://localhost/db" {
		t.Errorf("unexpected DATABASE_URL: %q", m["DATABASE_URL"])
	}
}

func TestCompareMissing(t *testing.T) {
	refs := map[string][]string{
		"DATABASE_URL": {"app.py"},
		"SECRET_KEY":   {"app.py"},
		"REDIS_HOST":   {"server.js"},
	}
	sources := map[string]map[string]string{
		".env": {"DATABASE_URL": "x", "SECRET_KEY": "y"},
	}
	result := Compare(refs, sources)
	if _, ok := result.Missing["REDIS_HOST"]; !ok {
		t.Error("expected REDIS_HOST to be missing")
	}
	if _, ok := result.Missing["DATABASE_URL"]; ok {
		t.Error("DATABASE_URL should not be missing")
	}
}

func TestParseDockerCompose(t *testing.T) {
	f := filepath.Join(t.TempDir(), "docker-compose.yml")
	os.WriteFile(f, []byte("services:\n  web:\n    environment:\n      - DATABASE_URL=postgres://db\n      - REDIS_HOST=redis\n  worker:\n    environment:\n      API_KEY: secret123\n"), 0644)
	m := ParseDockerCompose(f)
	if m["DATABASE_URL"] != "postgres://db" {
		t.Errorf("expected DATABASE_URL=postgres://db, got %q", m["DATABASE_URL"])
	}
	if m["API_KEY"] != "secret123" {
		t.Errorf("expected API_KEY=secret123, got %q", m["API_KEY"])
	}
	if len(m) != 3 {
		t.Errorf("expected 3 vars, got %d: %v", len(m), m)
	}
}

func TestDriftDetection(t *testing.T) {
	refs := map[string][]string{"APP_NAME": {"app.go"}}
	sources := map[string]map[string]string{
		".env":           {"APP_NAME": "x", "DB_HOST": "y"},
		"docker-compose": {"APP_NAME": "x"},
	}
	result := Compare(refs, sources)
	found := false
	for _, d := range result.Drift {
		if d.Var == "DB_HOST" {
			found = true
			if len(d.In) != 1 || d.In[0] != ".env" {
				t.Errorf("expected DB_HOST in .env only, got %v", d.In)
			}
			if len(d.NotIn) != 1 || d.NotIn[0] != "docker-compose" {
				t.Errorf("expected DB_HOST not in docker-compose, got %v", d.NotIn)
			}
		}
	}
	if !found {
		t.Error("expected drift entry for DB_HOST")
	}
}

func TestParseK8sConfigMap(t *testing.T) {
	f := filepath.Join(t.TempDir(), "configmap.yaml")
	os.WriteFile(f, []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: app-config\ndata:\n  DB_HOST: postgres\n  CACHE_TTL: \"300\"\n"), 0644)
	m := ParseK8sConfigMap(f)
	if m["DB_HOST"] != "postgres" {
		t.Errorf("expected DB_HOST=postgres, got %q", m["DB_HOST"])
	}
	if m["CACHE_TTL"] != "300" {
		t.Errorf("expected CACHE_TTL=300, got %q", m["CACHE_TTL"])
	}
}
