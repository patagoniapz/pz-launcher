// Package manifest define el formato del manifest.json y su descarga.
//
// El manifiesto es el "contrato" entre el servidor (GitHub Releases) y el
// launcher: describe qué ficheros existen, su hash SHA-256 esperado, de dónde
// bajarlos, qué versión del propio launcher es la actual y cómo arrancar el
// juego en cada plataforma.
package manifest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Manifest es el documento raíz publicado en el release.
type Manifest struct {
	// Schema versiona el formato del propio manifiesto por si evoluciona.
	Schema int `json:"schema"`

	// LauncherVersion es la última versión disponible del launcher (semver).
	LauncherVersion string `json:"launcher_version"`

	// LauncherBinaries mapea "GOOS/GOARCH" -> binario del launcher, para la
	// autoactualización. Ej.: "windows/amd64", "linux/amd64", "darwin/arm64".
	LauncherBinaries map[string]Asset `json:"launcher_binaries"`

	// Files son los ficheros del juego gestionados (p. ej. projectzomboid.jar).
	Files []File `json:"files"`

	// Launch indica cómo arrancar el juego por SO (clave = runtime.GOOS).
	Launch map[string]LaunchSpec `json:"launch"`
}

// Asset es un fichero descargable con su hash de verificación.
type Asset struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// File es un fichero del juego gestionado por el launcher.
type File struct {
	// Path es relativo a la carpeta del ejecutable, con separador "/".
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	URL    string `json:"url"`
}

// LaunchSpec describe el ejecutable del juego a lanzar tras actualizar.
type LaunchSpec struct {
	// Exe es relativo a la carpeta del launcher (p. ej. "ProjectZomboid64.exe").
	Exe  string   `json:"exe"`
	Args []string `json:"args"`
}

// Fetch descarga y decodifica el manifiesto desde una URL.
func Fetch(url string) (*Manifest, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("descargando manifiesto: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifiesto devolvió HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20)) // 5 MiB de tope
	if err != nil {
		return nil, fmt.Errorf("leyendo manifiesto: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parseando manifiesto: %w", err)
	}
	return &m, nil
}
