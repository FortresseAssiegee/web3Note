// This tool builds a local-only module proxy for a controlled MVS exercise.
package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type release struct {
	module, version, gomod string
	files                  map[string]string
}

func write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func publish(root string, r release) error {
	dir := filepath.Join(root, filepath.FromSlash(r.module), "@v")
	info, err := json.Marshal(struct {
		Version string
		Time    string
	}{r.version, "2024-01-01T00:00:00Z"})
	if err != nil {
		return err
	}
	if err := write(filepath.Join(dir, r.version+".info"), info); err != nil {
		return err
	}
	if err := write(filepath.Join(dir, r.version+".mod"), []byte(r.gomod)); err != nil {
		return err
	}
	all := map[string]string{"go.mod": r.gomod}
	for name, source := range r.files {
		all[name] = source
	}
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)
	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	for _, name := range names {
		f, err := z.Create(r.module + "@" + r.version + "/" + name)
		if err != nil {
			return err
		}
		if _, err := f.Write([]byte(all[name])); err != nil {
			return err
		}
	}
	if err := z.Close(); err != nil {
		return err
	}
	return write(filepath.Join(dir, r.version+".zip"), buf.Bytes())
}

func main() {
	out := flag.String("out", "proxy", "new local proxy directory")
	flag.Parse()
	releases := []release{
		{"example.com/learning/number", "v1.1.0", "module example.com/learning/number\n\ngo 1.24.0\n", map[string]string{"number.go": "package number\nfunc Value() int { return 11 }\n"}},
		{"example.com/learning/number", "v1.3.0", "module example.com/learning/number\n\ngo 1.24.0\n", map[string]string{"number.go": "package number\nfunc Value() int { return 13 }\n"}},
		{"example.com/learning/left", "v1.0.0", "module example.com/learning/left\n\ngo 1.24.0\n\nrequire example.com/learning/number v1.1.0\n", map[string]string{"left.go": "package left\nimport \"example.com/learning/number\"\nfunc Value() int { return number.Value() }\n"}},
		{"example.com/learning/right", "v1.0.0", "module example.com/learning/right\n\ngo 1.24.0\n\nrequire example.com/learning/number v1.3.0\n", map[string]string{"right.go": "package right\nimport \"example.com/learning/number\"\nfunc Value() int { return number.Value() }\n"}},
	}
	versions := make(map[string][]string)
	for _, r := range releases {
		if err := publish(*out, r); err != nil {
			panic(err)
		}
		versions[r.module] = append(versions[r.module], r.version)
	}
	for module, list := range versions {
		sort.Strings(list)
		if err := write(filepath.Join(*out, filepath.FromSlash(module), "@v", "list"), []byte(strings.Join(list, "\n")+"\n")); err != nil {
			panic(err)
		}
	}
	fmt.Println("local releases: 4")
}
