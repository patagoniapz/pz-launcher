// Package config lee/escribe un fichero de configuración opcional situado junto
// al ejecutable (pzlauncher.json). Permite al usuario personalizar cosas como la
// ruta de la carpeta de datos de Zomboid, que normalmente se autodetecta.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config son los ajustes persistentes del launcher.
type Config struct {
	// ZomboidDir sobrescribe la carpeta de datos de Zomboid autodetectada
	// (%UserProfile%\Zomboid). Vacío = autodetectar.
	ZomboidDir string `json:"zomboid_dir,omitempty"`
}

func file(baseDir string) string {
	return filepath.Join(baseDir, "pzlauncher.json")
}

// Load lee la config junto al ejecutable. Si no existe o está mal formada,
// devuelve una config vacía (comportamiento por defecto).
func Load(baseDir string) Config {
	var c Config
	data, err := os.ReadFile(file(baseDir))
	if err != nil {
		return c
	}
	_ = json.Unmarshal(data, &c)
	return c
}

// Save escribe la config junto al ejecutable.
func Save(baseDir string, c Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file(baseDir), append(data, '\n'), 0o644)
}
