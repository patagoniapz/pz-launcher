//go:build windows

// Package sysmem detecta la RAM física total del sistema. Se usa para sugerir
// un valor de -Xmx razonable al jugador en los ajustes de rendimiento.
package sysmem

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// memoryStatusEx refleja la estructura MEMORYSTATUSEX de la WinAPI.
type memoryStatusEx struct {
	length               uint32
	memoryLoad           uint32
	totalPhys            uint64
	availPhys            uint64
	totalPageFile        uint64
	availPageFile        uint64
	totalVirtual         uint64
	availVirtual         uint64
	availExtendedVirtual uint64
}

// TotalBytes devuelve la RAM física total en bytes, o 0 si no se pudo leer.
func TotalBytes() uint64 {
	proc := windows.NewLazySystemDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
	var m memoryStatusEx
	m.length = uint32(unsafe.Sizeof(m))
	if r, _, _ := proc.Call(uintptr(unsafe.Pointer(&m))); r == 0 {
		return 0
	}
	return m.totalPhys
}
