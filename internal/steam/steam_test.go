package steam

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGameInLibrary(t *testing.T) {
	lib := t.TempDir()
	// Estructura: <lib>/steamapps/appmanifest_108600.acf + common/ProjectZomboid
	acf := filepath.Join(lib, "steamapps", "appmanifest_108600.acf")
	mustWrite(t, acf, `"AppState"
{
	"appid"  "108600"
	"installdir"  "ProjectZomboid"
}`)
	gameDir := filepath.Join(lib, "steamapps", "common", "ProjectZomboid")
	if err := os.MkdirAll(gameDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got, ok := gameInLibrary(lib, 108600)
	if !ok {
		t.Fatal("no encontró el juego en la biblioteca")
	}
	if got != gameDir {
		t.Fatalf("ruta incorrecta: %s (esperaba %s)", got, gameDir)
	}

	// AppID que no existe.
	if _, ok := gameInLibrary(lib, 999999); ok {
		t.Fatal("no debería encontrar un AppID inexistente")
	}
}

func TestLibraryPaths(t *testing.T) {
	root := t.TempDir()
	vdf := filepath.Join(root, "steamapps", "libraryfolders.vdf")
	mustWrite(t, vdf, `"libraryfolders"
{
	"0"
	{
		"path"  "C:\\Program Files (x86)\\Steam"
	}
	"1"
	{
		"path"  "D:\\SteamLibrary"
	}
}`)

	libs := libraryPaths(root)
	// Debe incluir el propio root y las dos rutas (desescapadas).
	wantContains := []string{root, `C:\Program Files (x86)\Steam`, `D:\SteamLibrary`}
	for _, w := range wantContains {
		found := false
		for _, l := range libs {
			if l == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("falta la biblioteca %q en %v", w, libs)
		}
	}
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
