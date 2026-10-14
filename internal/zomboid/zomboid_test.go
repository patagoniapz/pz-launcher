package zomboid

import (
	"archive/zip"
	"fmt"
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
	mustWrite(t, filepath.Join(dir, "logs.zip"), "paquete de logs")
	// Cosas que NO se deben tocar:
	mustWrite(t, filepath.Join(dir, "backup.zip"), "copia del usuario")
	mustWrite(t, filepath.Join(dir, "Saves", "partida", "map.bin"), "partida")
	mustWrite(t, filepath.Join(dir, "mods", "mimod", "mod.info"), "mod")
	mustWrite(t, filepath.Join(dir, "options.ini"), "opciones")

	removed, err := CleanLogs(dir)
	if err != nil {
		t.Fatalf("CleanLogs error: %v", err)
	}
	if removed != 5 {
		t.Fatalf("esperaba borrar 5 elementos, borró %d", removed)
	}

	// Los logs deben haber desaparecido.
	assertGone(t, filepath.Join(dir, "Logs", "01-10-26_DebugLog.txt"))
	assertGone(t, filepath.Join(dir, "console.txt"))
	assertGone(t, filepath.Join(dir, "coop-console.txt"))
	assertGone(t, filepath.Join(dir, "logs.zip"))
	// La carpeta Logs se conserva (vacía).
	assertExists(t, filepath.Join(dir, "Logs"))
	// Partidas, mods, opciones y otros zips del usuario intactos.
	assertExists(t, filepath.Join(dir, "backup.zip"))
	assertExists(t, filepath.Join(dir, "Saves", "partida", "map.bin"))
	assertExists(t, filepath.Join(dir, "mods", "mimod", "mod.info"))
	assertExists(t, filepath.Join(dir, "options.ini"))
}

func TestBackupLogs(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "Logs", "2026-10-14_11-57_DebugLog.txt"), "debug")
	mustWrite(t, filepath.Join(dir, "Logs", "2026-10-14_11-57_client chat Bob.txt"), "chat")
	mustWrite(t, filepath.Join(dir, "console.txt"), "consola")
	// Cosas que NO deben entrar al zip:
	mustWrite(t, filepath.Join(dir, "options.ini"), "opciones")
	mustWrite(t, filepath.Join(dir, "Saves", "p", "map.bin"), "partida")

	zipPath, n, err := BackupLogs(dir)
	if err != nil {
		t.Fatalf("BackupLogs error: %v", err)
	}
	if n != 3 {
		t.Fatalf("esperaba respaldar 3 ficheros, respaldó %d", n)
	}
	// El zip vive en la subcarpeta patagoniapz/.
	if filepath.Base(filepath.Dir(zipPath)) != backupSubdir {
		t.Fatalf("el zip no quedó en %s: %s", backupSubdir, zipPath)
	}
	// Contenido esperado del zip.
	want := map[string]bool{
		"Logs/2026-10-14_11-57_DebugLog.txt":        true,
		"Logs/2026-10-14_11-57_client chat Bob.txt": true,
		"console.txt": true,
	}
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("abriendo zip: %v", err)
	}
	defer zr.Close()
	got := map[string]bool{}
	for _, f := range zr.File {
		got[f.Name] = true
	}
	for name := range want {
		if !got[name] {
			t.Errorf("falta %q en el zip; contenido: %v", name, got)
		}
	}
	if got["options.ini"] || len(got) != 3 {
		t.Errorf("el zip contiene ficheros de más: %v", got)
	}

	// Carpeta vacía -> no crea zip.
	empty := t.TempDir()
	zp, n2, err := BackupLogs(empty)
	if err != nil || zp != "" || n2 != 0 {
		t.Fatalf("sin logs no debía crear zip: zp=%q n=%d err=%v", zp, n2, err)
	}
}

func TestPruneBackups(t *testing.T) {
	dir := t.TempDir()
	// 13 zips nuestros (con fecha ordenable) + 1 zip ajeno del usuario.
	for i := 1; i <= 13; i++ {
		mustWrite(t, filepath.Join(dir, fmt.Sprintf("%s2026-10-%02d_00-00-00.zip", backupPrefix, i)), "z")
	}
	mustWrite(t, filepath.Join(dir, "mi-backup-personal.zip"), "ajeno")

	pruneBackups(dir, backupsKeep)

	entries, _ := os.ReadDir(dir)
	ours := 0
	for _, en := range entries {
		if len(en.Name()) >= len(backupPrefix) && en.Name()[:len(backupPrefix)] == backupPrefix {
			ours++
		}
	}
	if ours != backupsKeep {
		t.Fatalf("esperaba conservar %d respaldos, quedaron %d", backupsKeep, ours)
	}
	// El más viejo (día 01) debió borrarse; el más nuevo (día 13) debe seguir.
	assertGone(t, filepath.Join(dir, backupPrefix+"2026-10-01_00-00-00.zip"))
	assertExists(t, filepath.Join(dir, backupPrefix+"2026-10-13_00-00-00.zip"))
	// El zip ajeno del usuario NO se toca.
	assertExists(t, filepath.Join(dir, "mi-backup-personal.zip"))
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
