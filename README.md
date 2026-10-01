# pzlauncher

> 🇬🇧 English below · 🇪🇸 Español más abajo

---

## 🇬🇧 English

Cross-platform **launcher / updater for Project Zomboid**. A single executable
you drop **inside the game folder** (e.g.
`...\Steam\steamapps\common\ProjectZomboid`). On launch it:

1. Reads a `manifest.json` from GitHub Releases.
2. Self-updates if a newer launcher version is published.
3. Compares each managed file (e.g. `projectzomboid.jar`) by **SHA-256** and
   downloads only what changed.
4. Starts the game.

### Get it

A prebuilt binary is provided in the [Releases](../../releases) section for
Windows, Linux and macOS — just download and drop it in the game folder.

If you prefer, you can **download the source and build it yourself**:

```bash
# requires Go 1.22+
git clone https://github.com/patagoniapz/pz-launcher.git
cd pz-launcher
go build -o pzlauncher ./cmd/launcher     # add .exe on Windows
```

Optionally pin version and manifest URL:

```bash
go build -ldflags "-X main.version=1.0.0 -X main.manifestURL=https://github.com/patagoniapz/pz-launcher/releases/latest/download/manifest.json" -o pzlauncher ./cmd/launcher
```

---

## 🇪🇸 Español

**Launcher / actualizador multiplataforma para Project Zomboid**. Un único
ejecutable que se deja **dentro de la carpeta del juego** (p. ej.
`...\Steam\steamapps\common\ProjectZomboid`). Al arrancar:

1. Lee un `manifest.json` desde GitHub Releases.
2. Se autoactualiza si hay una versión más nueva del launcher.
3. Compara cada fichero gestionado (p. ej. `projectzomboid.jar`) por **SHA-256**
   y descarga solo lo que haya cambiado.
4. Arranca el juego.

### Cómo obtenerlo

Se ofrece un binario ya compilado en la sección de [Releases](../../releases)
para Windows, Linux y macOS — basta con descargarlo y dejarlo en la carpeta del
juego.

Si lo prefieres, puedes **descargar el código fuente y compilarlo por tus
propios medios**:

```bash
# requiere Go 1.22+
git clone https://github.com/patagoniapz/pz-launcher.git
cd pz-launcher
go build -o pzlauncher ./cmd/launcher     # añade .exe en Windows
```

Opcionalmente, fija la versión y la URL del manifiesto:

```bash
go build -ldflags "-X main.version=1.0.0 -X main.manifestURL=https://github.com/patagoniapz/pz-launcher/releases/latest/download/manifest.json" -o pzlauncher ./cmd/launcher
```
