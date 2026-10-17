package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"pzlauncher/internal/manifest"
)

// nopReporter implementa ui.Reporter sin hacer nada (para los tests).
type nopReporter struct{}

func (nopReporter) Stage(string)         {}
func (nopReporter) Progress(_, _ int64)  {}
func (nopReporter) Success(string)       {}
func (nopReporter) Close()               {}

func sha256hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// serveContent levanta un servidor HTTP de prueba que devuelve content.
func serveContent(t *testing.T, content []byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
}

func TestSyncFiles_DownloadsWhenMissing(t *testing.T) {
	content := []byte("contenido del projectzomboid.jar")
	srv := serveContent(t, content)
	defer srv.Close()

	base := t.TempDir()
	files := []manifest.File{{
		Path:   "projectzomboid.jar",
		SHA256: sha256hex(content),
		Size:   int64(len(content)),
		URL:    srv.URL + "/projectzomboid.jar",
	}}

	if err := SyncFiles(base, files, nopReporter{}); err != nil {
		t.Fatalf("SyncFiles devolvió error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(base, "projectzomboid.jar"))
	if err != nil {
		t.Fatalf("no se escribió el fichero: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("contenido incorrecto: %q", got)
	}
}

func TestSyncFiles_SkipsWhenUpToDate(t *testing.T) {
	content := []byte("ya actualizado")
	// El servidor falla el test si se le pide descargar.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no debería descargar un fichero ya actualizado")
	}))
	defer srv.Close()

	base := t.TempDir()
	dest := filepath.Join(base, "projectzomboid.jar")
	if err := os.WriteFile(dest, content, 0o644); err != nil {
		t.Fatal(err)
	}

	files := []manifest.File{{
		Path:   "projectzomboid.jar",
		SHA256: sha256hex(content),
		URL:    srv.URL + "/projectzomboid.jar",
	}}

	if err := SyncFiles(base, files, nopReporter{}); err != nil {
		t.Fatalf("SyncFiles devolvió error: %v", err)
	}
}

func TestResolveDest_WindowsUsesRoot(t *testing.T) {
	base := t.TempDir()
	got := resolveDest("windows", base, "projectzomboid.jar")
	want := filepath.Join(base, "projectzomboid.jar")
	if got != want {
		t.Fatalf("resolveDest = %q, se esperaba %q", got, want)
	}
}

func TestResolveDest_DarwinPrefersExistingBundleCopy(t *testing.T) {
	base := t.TempDir()
	javaDir := filepath.Join(base, "Project Zomboid.app", "Contents", "Java")
	if err := os.MkdirAll(javaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	jar := filepath.Join(javaDir, "projectzomboid.jar")
	if err := os.WriteFile(jar, []byte("jar del bundle"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := resolveDest("darwin", base, "projectzomboid.jar"); got != jar {
		t.Fatalf("resolveDest = %q, se esperaba el jar del bundle %q", got, jar)
	}
}

func TestResolveDest_DarwinPlacesNewFileInsideBundle(t *testing.T) {
	base := t.TempDir()
	javaDir := filepath.Join(base, "Project Zomboid.app", "Contents", "Java")
	if err := os.MkdirAll(javaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// El jar no existe todavía en ningún candidato: debe ir al bundle porque
	// su carpeta ya existe.
	want := filepath.Join(javaDir, "projectzomboid.jar")
	if got := resolveDest("darwin", base, "projectzomboid.jar"); got != want {
		t.Fatalf("resolveDest = %q, se esperaba %q", got, want)
	}
}

func TestResolveDest_DarwinWithoutBundleFallsBackToRoot(t *testing.T) {
	base := t.TempDir()
	want := filepath.Join(base, "projectzomboid.jar")
	if got := resolveDest("darwin", base, "projectzomboid.jar"); got != want {
		t.Fatalf("resolveDest = %q, se esperaba la raíz %q", got, want)
	}
}

func TestResolveDest_LinuxPrefersExistingCopyInSubdir(t *testing.T) {
	base := t.TempDir()
	sub := filepath.Join(base, "projectzomboid")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	jar := filepath.Join(sub, "projectzomboid.jar")
	if err := os.WriteFile(jar, []byte("jar en subcarpeta"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := resolveDest("linux", base, "projectzomboid.jar"); got != jar {
		t.Fatalf("resolveDest = %q, se esperaba %q", got, jar)
	}
	// Y si existe en la raíz, la raíz tiene prioridad (orden de candidatos).
	root := filepath.Join(base, "projectzomboid.jar")
	if err := os.WriteFile(root, []byte("jar en la raíz"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := resolveDest("linux", base, "projectzomboid.jar"); got != root {
		t.Fatalf("resolveDest = %q, se esperaba %q", got, root)
	}
}

func TestSyncFiles_FailsOnHashMismatch(t *testing.T) {
	content := []byte("contenido real")
	srv := serveContent(t, content)
	defer srv.Close()

	base := t.TempDir()
	files := []manifest.File{{
		Path:   "projectzomboid.jar",
		SHA256: sha256hex([]byte("otro contenido distinto")), // hash que NO coincide
		URL:    srv.URL + "/projectzomboid.jar",
	}}

	if err := SyncFiles(base, files, nopReporter{}); err == nil {
		t.Fatal("se esperaba error por hash que no coincide, pero fue nil")
	}
	// No debe quedar el fichero corrupto en su sitio.
	if _, err := os.Stat(filepath.Join(base, "projectzomboid.jar")); !os.IsNotExist(err) {
		t.Fatal("no debería existir el fichero tras un fallo de hash")
	}
}
