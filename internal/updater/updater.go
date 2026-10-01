// Package updater sincroniza los ficheros del juego gestionados contra el
// manifiesto: compara hashes y reemplaza solo lo que haya cambiado.
package updater

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"pzlauncher/internal/download"
	"pzlauncher/internal/hashutil"
	"pzlauncher/internal/manifest"
)

// SyncFiles recorre los ficheros del manifiesto, comparando el hash local con
// el esperado. Descarga y reemplaza únicamente los que falten o difieran.
// baseDir es la carpeta del launcher; los paths del manifiesto son relativos a ella.
func SyncFiles(baseDir string, files []manifest.File) error {
	for _, f := range files {
		dest := filepath.Join(baseDir, filepath.FromSlash(f.Path))

		local, err := hashutil.SHA256File(dest)
		if err != nil {
			return fmt.Errorf("hash local de %s: %w", f.Path, err)
		}
		if local == f.SHA256 {
			log.Printf("[ok] %s está actualizado", f.Path)
			continue
		}

		if local == "" {
			log.Printf("[nuevo] %s no existe, descargando...", f.Path)
		} else {
			log.Printf("[actualizar] %s cambió, descargando...", f.Path)
		}

		if err := replaceFile(baseDir, dest, f); err != nil {
			return err
		}
		log.Printf("[listo] %s actualizado", f.Path)
	}
	return nil
}

// replaceFile descarga el fichero a un temporal (verificando su hash), hace una
// copia de seguridad .bak del actual si existía, y mueve el nuevo a su sitio.
func replaceFile(baseDir, dest string, f manifest.File) error {
	// El temporal se crea en la carpeta destino para que os.Rename sea atómico
	// (mismo volumen) y no un copy-across-devices.
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("creando carpeta de %s: %w", f.Path, err)
	}

	tmpPath, err := download.Verified(f.URL, filepath.Dir(dest), f.SHA256)
	if err != nil {
		return err
	}

	bak := dest + ".bak"
	if _, err := os.Stat(dest); err == nil {
		_ = os.Remove(bak)
		if err := os.Rename(dest, bak); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("respaldando %s: %w", f.Path, err)
		}
	}

	if err := os.Rename(tmpPath, dest); err != nil {
		// Intentar restaurar el respaldo si el swap falla.
		_ = os.Rename(bak, dest)
		os.Remove(tmpPath)
		return fmt.Errorf("reemplazando %s: %w", f.Path, err)
	}

	// Éxito: el .bak queda como red de seguridad hasta la próxima actualización.
	return nil
}
