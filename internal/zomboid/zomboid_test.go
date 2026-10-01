package zomboid

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanLogs(t *testing.T) {
	dir := t.TempDir()

	// Estructura simulada de %UserProfile%\Zomboid
	mustWrite(t, filepath.Join(dir, "Logs", "01-10-26_DebugLog.txt"), "log")
	mustWrite(t, filepath.Join(dir, "Logs", "01-10-26_server.txt"), "log")
	mustWrite(t, filepath.Join(dir, "console.txt"), "consola")
	mustWrite(t, filepath.Join(dir, "coop-console.txt"), "consola coop")
	// Cosas que NO se deben tocar:
	mustWrite(t, filepath.Join(dir, "Saves", "partida", "map.bin"), "partida")
	mustWrite(t, filepath.Join(dir, "mods", "mimod", "mod.info"), "mod")
	mustWrite(t, filepath.Join(dir, "options.ini"), "opciones")

	removed, err := CleanLogs(dir)
	if err != nil {
		t.Fatalf("CleanLogs error: %v", err)
	}
	if removed != 4 {
		t.Fatalf("esperaba borrar 4 elementos, borró %d", removed)
	}

	// Los logs deben haber desaparecido.
	assertGone(t, filepath.Join(dir, "Logs", "01-10-26_DebugLog.txt"))
	assertGone(t, filepath.Join(dir, "console.txt"))
	assertGone(t, filepath.Join(dir, "coop-console.txt"))
	// La carpeta Logs se conserva (vacía).
	assertExists(t, filepath.Join(dir, "Logs"))
	// Partidas, mods y opciones intactos.
	assertExists(t, filepath.Join(dir, "Saves", "partida", "map.bin"))
	assertExists(t, filepath.Join(dir, "mods", "mimod", "mod.info"))
	assertExists(t, filepath.Join(dir, "options.ini"))
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertGone(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("debería haberse borrado: %s", path)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("debería seguir existiendo: %s (%v)", path, err)
	}
}
