//go:build !windows

package steam

import (
	"os"
	"path/filepath"
)

// steamRoots devuelve las ubicaciones habituales de Steam en Linux y macOS.
func steamRoots() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(home, ".steam", "steam"),              // Linux (symlink)
		filepath.Join(home, ".local", "share", "Steam"),     // Linux
		filepath.Join(home, ".steam", "root"),               // Linux (variante)
		filepath.Join(home, "Library", "Application Support", "Steam"), // macOS
	}
}
