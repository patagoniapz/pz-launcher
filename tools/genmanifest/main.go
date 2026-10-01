// Command genmanifest genera el manifest.json a partir de los ficheros locales:
// calcula su SHA-256 y tamaño y construye la URL de descarga en el release.
//
// Ejemplo (lo usa la CI tras compilar los binarios):
//
//	genmanifest \
//	  -release-url https://github.com/OWNER/REPO/releases/download/v1.0.0 \
//	  -version 1.0.0 \
//	  -game projectzomboid.jar=dist/projectzomboid.jar \
//	  -bin windows/amd64=dist/pzlauncher-windows-amd64.exe \
//	  -bin linux/amd64=dist/pzlauncher-linux-amd64 \
//	  -bin darwin/arm64=dist/pzlauncher-darwin-arm64 \
//	  -out dist/manifest.json
//
// En -game y -bin, la parte izquierda del "=" es la clave en el manifiesto
// (ruta relativa del fichero del juego, o "GOOS/GOARCH") y la derecha es el
// fichero local del que se calcula el hash. La URL de descarga se forma como
// release-url + "/" + nombre-del-fichero (GitHub aplana los assets por nombre).
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"pzlauncher/internal/manifest"
)

type kvFlag []string

func (k *kvFlag) String() string { return strings.Join(*k, ",") }
func (k *kvFlag) Set(v string) error {
	*k = append(*k, v)
	return nil
}

func main() {
	var (
		releaseURL = flag.String("release-url", "", "URL base del release (sin barra final)")
		gameFiles  kvFlag
		binFiles   kvFlag
		version    = flag.String("version", "", "versión del launcher (semver)")
		out        = flag.String("out", "manifest.json", "ruta de salida del manifiesto")
		launchWin  = flag.String("launch-windows", "ProjectZomboid64.exe", "ejecutable del juego en Windows")
		launchLin  = flag.String("launch-linux", "./ProjectZomboid64", "ejecutable del juego en Linux")
		launchMac  = flag.String("launch-darwin", "./ProjectZomboid64", "ejecutable del juego en macOS")
	)
	flag.Var(&gameFiles, "game", "fichero del juego: rutaRelativa=ficheroLocal (repetible)")
	flag.Var(&binFiles, "bin", "binario del launcher: GOOS/GOARCH=ficheroLocal (repetible)")
	flag.Parse()

	if *releaseURL == "" {
		fail("falta -release-url")
	}
	base := strings.TrimRight(*releaseURL, "/")

	m := manifest.Manifest{
		Schema:           1,
		LauncherVersion:  *version,
		LauncherBinaries: map[string]manifest.Asset{},
		Launch: map[string]manifest.LaunchSpec{
			"windows": {Exe: *launchWin},
			"linux":   {Exe: *launchLin},
			"darwin":  {Exe: *launchMac},
		},
	}

	for _, entry := range gameFiles {
		relPath, local := split(entry)
		sum, size := hashAndSize(local)
		m.Files = append(m.Files, manifest.File{
			Path:   filepath.ToSlash(relPath),
			SHA256: sum,
			Size:   size,
			URL:    base + "/" + filepath.Base(local),
		})
	}

	for _, entry := range binFiles {
		key, local := split(entry)
		sum, size := hashAndSize(local)
		m.LauncherBinaries[key] = manifest.Asset{
			URL:    base + "/" + filepath.Base(local),
			SHA256: sum,
			Size:   size,
		}
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fail("serializando manifiesto: %v", err)
	}
	if err := os.WriteFile(*out, append(data, '\n'), 0o644); err != nil {
		fail("escribiendo %s: %v", *out, err)
	}
	fmt.Printf("manifiesto escrito en %s (%d ficheros, %d binarios)\n", *out, len(m.Files), len(m.LauncherBinaries))
}

func split(entry string) (key, local string) {
	i := strings.Index(entry, "=")
	if i < 0 {
		fail("entrada inválida %q, se esperaba clave=ficheroLocal", entry)
	}
	return entry[:i], entry[i+1:]
}

func hashAndSize(path string) (string, int64) {
	f, err := os.Open(path)
	if err != nil {
		fail("abriendo %s: %v", path, err)
	}
	defer f.Close()

	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		fail("leyendo %s: %v", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), n
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "genmanifest: "+format+"\n", args...)
	os.Exit(1)
}
