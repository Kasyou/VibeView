package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ProjectType string

const (
	React   ProjectType = "react"
	Vue     ProjectType = "vue"
	Svelte  ProjectType = "svelte"
	HTML    ProjectType = "html"
	Unknown ProjectType = "unknown"
)

type ProjectInfo struct {
	Type          ProjectType
	DevServerPort int
	DevServerURL  string
	WatchDirs     []string
	ServeLocal    bool // true for HTML projects (VibeView serves the files)
}

func Detect(root string) ProjectInfo {
	info := ProjectInfo{
		Type:          Unknown,
		DevServerPort: 5173,
	}

	pkg := readPackageJSON(root)
	hasVite := viteConfigExists(root)
	hasHTML := fileExists(filepath.Join(root, "index.html"))

	// Try framework detection by Vite plugin first (most reliable),
	// then by framework dependency, then by project config file.
	switch {
	// React: @vitejs/plugin-react or react dep with vite
	case pkg.hasDep("@vitejs/plugin-react") || (pkg.hasDep("react") && hasVite):
		info.Type = React
		info.WatchDirs = dirs(root, "src")
		info.DevServerURL = "http://localhost:5173"

	// Vue: @vitejs/plugin-vue or vue dep with vite
	case pkg.hasDep("@vitejs/plugin-vue") || (pkg.hasDep("vue") && hasVite):
		info.Type = Vue
		info.WatchDirs = dirs(root, "src")
		info.DevServerURL = "http://localhost:5173"

	// Svelte: @sveltejs/vite-plugin-svelte, svelte dep, or SvelteKit
	case pkg.hasDep("@sveltejs/vite-plugin-svelte") ||
		pkg.hasDep("@sveltejs/kit") ||
		pkg.hasDep("svelte"):
		info.Type = Svelte
		info.WatchDirs = dirs(root, "src")
		info.DevServerURL = "http://localhost:5173"

	// HTML: has an index.html
	case hasHTML:
		info.Type = HTML
		info.WatchDirs = []string{root}
		info.ServeLocal = true

	// Fallback: treat as HTML project (serve files locally)
	default:
		info.Type = HTML
		info.WatchDirs = []string{root}
		info.ServeLocal = true
	}

	return info
}

type pkgJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func readPackageJSON(root string) pkgJSON {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return pkgJSON{}
	}
	var p pkgJSON
	json.Unmarshal(data, &p)
	return p
}

func (p pkgJSON) hasDep(name string) bool {
	if _, ok := p.Dependencies[name]; ok {
		return true
	}
	if _, ok := p.DevDependencies[name]; ok {
		return true
	}
	return false
}

func viteConfigExists(root string) bool {
	return fileExists(filepath.Join(root, "vite.config.js")) ||
		fileExists(filepath.Join(root, "vite.config.ts")) ||
		fileExists(filepath.Join(root, "vite.config.mjs"))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirs(root string, names ...string) []string {
	var result []string
	for _, n := range names {
		p := filepath.Join(root, n)
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		result = append(result, root)
	}
	return result
}
