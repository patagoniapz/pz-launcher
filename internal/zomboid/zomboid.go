// Package zomboid localiza la carpeta de datos del juego (logs, partidas…) y
// limpia los logs del cliente.
//
// Ojo: esta carpeta NO es la del launcher. El launcher vive en la carpeta de
// instalación (…\steamapps\common\ProjectZomboid), mientras que los datos del
// usuario están en %UserProfile%\Zomboid (equivalente a ~/Zomboid en Linux/Mac).
package zomboid

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// backupSubdir es la subcarpeta (en la raíz de datos de Zomboid) donde dejamos
// los .zip de respaldo de logs, uno por sesión con la fecha en el nombre.
const backupSubdir = "patagoniapz"

// backupsKeep es cuántos .zip de respaldo conservamos; los más viejos se borran.
const backupsKeep = 10

// backupPrefix es el prefijo del nombre de los zips que gestionamos (para la
// poda: no tocamos ningún otro .zip que el usuario tenga ahí).
const backupPrefix = "logs-backup_"

// DefaultDir devuelve la ruta estándar de datos de Zomboid: <home>/Zomboid.
func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Zomboid")
}

// Resolve decide la carpeta de datos de Zomboid con esta prioridad:
//  1. Variable de entorno PZL_ZOMBOID_DIR.
//  2. Valor configurado por el usuario (pzlauncher.json).
//  3. Autodetección (<home>/Zomboid).
func Resolve(configured string) string {
	if env := os.Getenv("PZL_ZOMBOID_DIR"); env != "" {
		return env
	}
	if configured != "" {
		return configured
	}
	return DefaultDir()
}

// Exists indica si la carpeta de datos existe.
func Exists(dir string) bool {
	if dir == "" {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// BackupLogs comprime TODOS los logs del cliente (el contenido recursivo de
// <dir>/Logs y el console.txt de la raíz) en un .zip con fecha dentro de
// <dir>/patagoniapz, y conserva solo los backupsKeep más recientes. Pensado
// para correr al darle "Jugar", ANTES de CleanLogs, por si hay que revisar la
// sesión anterior. Devuelve la ruta del zip creado y cuántos ficheros metió.
// Best-effort: si no hay nada que respaldar devuelve ("", 0, nil) sin crear zip.
func BackupLogs(dir string) (zipPath string, count int, err error) {
	// 1. Reunir ficheros a respaldar: todo Logs/ (recursivo) + console.txt raíz.
	type item struct{ abs, rel string }
	var items []item
	logsDir := filepath.Join(dir, "Logs")
	_ = filepath.WalkDir(logsDir, func(p string, d os.DirEntry, werr error) error {
		if werr != nil || d.IsDir() {
			return nil
		}
		rel, e := filepath.Rel(dir, p)
		if e != nil {
			rel = filepath.Join("Logs", d.Name())
		}
		items = append(items, item{p, filepath.ToSlash(rel)})
		return nil
	})
	if console := filepath.Join(dir, "console.txt"); fileExists(console) {
		items = append(items, item{console, "console.txt"})
	}
	if len(items) == 0 {
		return "", 0, nil // no hay nada que respaldar
	}

	// 2. Crear el .zip con fecha en la subcarpeta de respaldos.
	backupDir := filepath.Join(dir, backupSubdir)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", 0, err
	}
	ts := time.Now().Format("2006-01-02_15-04-05")
	zipPath = filepath.Join(backupDir, backupPrefix+ts+".zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		return "", 0, err
	}
	zw := zip.NewWriter(zf)
	for _, it := range items {
		if addFileToZip(zw, it.abs, it.rel) == nil {
			count++
		}
	}
	// Cerrar el writer del zip ANTES que el fichero (flush del índice central).
	if cerr := zw.Close(); cerr != nil && err == nil {
		err = cerr
	}
	if cerr := zf.Close(); cerr != nil && err == nil {
		err = cerr
	}

	// 3. Retención: conservar solo los backupsKeep más recientes.
	pruneBackups(backupDir, backupsKeep)
	return zipPath, count, err
}

// addFileToZip agrega un fichero al zip bajo la ruta relativa indicada.
func addFileToZip(zw *zip.Writer, absPath, relPath string) error {
	src, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer src.Close()
	w, err := zw.Create(relPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, src)
	return err
}

// pruneBackups deja solo los `keep` zips de respaldo más recientes en backupDir
// (por nombre: el prefijo lleva la fecha YYYY-MM-DD_HH-MM-SS, que ordena
// cronológicamente). No toca ningún otro fichero del usuario.
func pruneBackups(backupDir string, keep int) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return
	}
	var zips []string
	for _, en := range entries {
		n := en.Name()
		if !en.IsDir() && strings.HasPrefix(n, backupPrefix) && strings.HasSuffix(strings.ToLower(n), ".zip") {
			zips = append(zips, n)
		}
	}
	if len(zips) <= keep {
		return
	}
	sort.Strings(zips)
	for _, n := range zips[:len(zips)-keep] {
		_ = os.Remove(filepath.Join(backupDir, n))
	}
}

// fileExists indica si path existe y es un fichero (no un directorio).
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// CleanLogs borra el contenido de <dir>/Logs y los ficheros de logs de la raíz
// (console.txt, *-console.txt y logs.zip). NO toca partidas, mods ni opciones.
// Devuelve cuántos elementos se eliminaron.
func CleanLogs(dir string) (removed int, err error) {
	// 1. Contenido de la carpeta Logs/ (se conserva la carpeta en sí).
	logsDir := filepath.Join(dir, "Logs")
	if entries, e := os.ReadDir(logsDir); e == nil {
		for _, en := range entries {
			if os.RemoveAll(filepath.Join(logsDir, en.Name())) == nil {
				removed++
			}
		}
	}

	// 2. Logs sueltos en la raíz de Zomboid: consolas y el logs.zip (el paquete
	// de logs comprimido que el juego deja junto a console.txt). Solo borramos
	// ESE zip concreto, no cualquier .zip del usuario.
	if entries, e := os.ReadDir(dir); e == nil {
		for _, en := range entries {
			if en.IsDir() {
				continue
			}
			name := en.Name()
			isConsole := name == "console.txt" || strings.HasSuffix(name, "-console.txt")
			isLogsZip := strings.EqualFold(name, "logs.zip")
			if isConsole || isLogsZip {
				if os.Remove(filepath.Join(dir, name)) == nil {
					removed++
				}
			}
		}
	}

	return removed, nil
}

// OpenDir abre dir en el explorador de archivos del sistema (cada usuario abre
// SU carpeta de datos de Zomboid, la misma que se usa para limpiar los logs).
func OpenDir(dir string) error {
	if dir == "" {
		return errors.New("ruta de carpeta vacía")
	}
	switch runtime.GOOS {
	case "windows":
		// explorer.exe suele devolver código de salida 1 aunque abra bien; por
		// eso lanzamos sin esperar (Start) y no interpretamos el resultado.
		return exec.Command("explorer", filepath.FromSlash(dir)).Start()
	case "darwin":
		return exec.Command("open", dir).Start()
	default:
		return exec.Command("xdg-open", dir).Start()
	}
}
