//go:build !windows

package sysmem

// TotalBytes no está implementado fuera de Windows (el cliente Patagonia es
// Windows). Devuelve 0 para que la sugerencia de RAM se omita de forma segura.
func TotalBytes() uint64 { return 0 }
