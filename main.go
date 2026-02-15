package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := flag.String("dir", ".", "directory to scan")
	envFile := flag.String("env", "", ".env file path")
	compose := flag.String("compose", "", "docker-compose.yml path")
	k8s := flag.String("k8s", "", "K8s ConfigMap YAML path")
	jsonOut := flag.Bool("json", false, "output as JSON")
	flag.Parse()

	refs := ScanCode(*dir)
	sources := map[string]map[string]string{}
	if *envFile != "" {
		sources[".env"] = ParseDotEnv(*envFile)
	}
	if *compose != "" {
		sources["docker-compose"] = ParseDockerCompose(*compose)
	}
	if *k8s != "" {
		sources["k8s-configmap"] = ParseK8sConfigMap(*k8s)
	}
	if len(sources) == 0 {
		if p := filepath.Join(*dir, ".env"); fileExists(p) {
			sources[".env"] = ParseDotEnv(p)
		}
		if p := filepath.Join(*dir, "docker-compose.yml"); fileExists(p) {
			sources["docker-compose"] = ParseDockerCompose(p)
		}
	}
	result := Compare(refs, sources)
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(result)
	} else {
		names := make([]string, 0, len(sources))
		for n := range sources {
			names = append(names, n)
		}
		fmt.Printf("\U0001F4E6 %d env vars in code | \U0001F4C2 %d sources: %s\n", len(refs), len(sources), strings.Join(names, ", "))
		if len(result.Missing) > 0 {
			fmt.Printf("\u274C %d vars missing:\n", len(result.Missing))
			for v, srcs := range result.Missing {
				fmt.Printf("  %-25s missing in: %s\n", v, strings.Join(srcs, ", "))
			}
		}
		if len(result.Drift) > 0 {
			fmt.Printf("\u26A0\uFE0F  %d vars drifting:\n", len(result.Drift))
			for _, d := range result.Drift {
				fmt.Printf("  %-25s in [%s] not in [%s]\n", d.Var, strings.Join(d.In, ","), strings.Join(d.NotIn, ","))
			}
		}
		if len(result.Missing) == 0 && len(result.Drift) == 0 {
			fmt.Println("\u2705 All environment variables verified!")
		}
	}
	if len(result.Missing) > 0 {
		os.Exit(1)
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
