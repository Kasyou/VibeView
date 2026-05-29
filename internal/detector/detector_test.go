package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectReact(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"dependencies":{"react":"^18.0.0"},"devDependencies":{"vite":"^5.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.ts"), []byte(`export default {}`), 0644)

	r := Detect(dir)
	if r.Type != React {
		t.Errorf("expected react, got %s", r.Type)
	}
}

func TestDetectVue(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"dependencies":{"vue":"^3.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte(`export default {}`), 0644)

	r := Detect(dir)
	if r.Type != Vue {
		t.Errorf("expected vue, got %s", r.Type)
	}
}

func TestDetectHTML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html></html>`), 0644)

	r := Detect(dir)
	if r.Type != HTML {
		t.Errorf("expected html, got %s", r.Type)
	}
}

func TestDetectSvelte(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"devDependencies":{"svelte":"^4.0.0","@sveltejs/vite-plugin-svelte":"^3.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte(`export default {}`), 0644)

	r := Detect(dir)
	if r.Type != Svelte {
		t.Errorf("expected svelte, got %s", r.Type)
	}
}

func TestDetectWatchDirs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"dependencies":{"react":"^18.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.ts"), []byte(`export default {}`), 0644)
	os.MkdirAll(filepath.Join(dir, "src"), 0755)

	r := Detect(dir)
	if len(r.WatchDirs) == 0 {
		t.Error("expected non-empty watch dirs")
	}
	if r.WatchDirs[0] != filepath.Join(dir, "src") {
		t.Errorf("expected src dir, got %s", r.WatchDirs[0])
	}
}

func TestDetectReactViaPlugin(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"devDependencies":{"@vitejs/plugin-react":"^4.0.0","vite":"^5.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte(`export default {}`), 0644)

	r := Detect(dir)
	if r.Type != React {
		t.Errorf("expected react via @vitejs/plugin-react, got %s", r.Type)
	}
}

func TestDetectVueViaPlugin(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"devDependencies":{"@vitejs/plugin-vue":"^5.0.0","vite":"^5.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte(`export default {}`), 0644)

	r := Detect(dir)
	if r.Type != Vue {
		t.Errorf("expected vue via @vitejs/plugin-vue, got %s", r.Type)
	}
}

func TestDetectSvelteKit(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"devDependencies":{"@sveltejs/kit":"^2.0.0"}}`), 0644)

	r := Detect(dir)
	if r.Type != Svelte {
		t.Errorf("expected svelte via @sveltejs/kit, got %s", r.Type)
	}
}

func TestDetectUnknownFallsToHTML(t *testing.T) {
	dir := t.TempDir()
	r := Detect(dir)
	if r.Type != HTML {
		t.Errorf("expected html for empty dir, got %s", r.Type)
	}
	if !r.ServeLocal {
		t.Error("empty dir should enable ServeLocal")
	}
}
