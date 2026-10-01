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

Progress is shown in a small **native window** with an **estimated time left**
(falls back to console if no GUI is available). On Windows it carries the
Project Zomboid icon.

It does **not** auto-launch: after updating it asks **Play / Clean logs & play /
Quit**. The game is started **through Steam** (`steam://rungameid/108600`), which
boots Steam if it is closed. "Clean logs" wipes `%UserProfile%\Zomboid\Logs\` and
the `*console.txt` files (never your saves or mods).

The Zomboid data folder is autodetected at `%UserProfile%\Zomboid` (`~/Zomboid`).
Override it with the `PZL_ZOMBOID_DIR` env var or a `pzlauncher.json` next to the
executable: `{ "zomboid_dir": "D:\\otra\\ruta\\Zomboid" }`.

### Get it

A prebuilt binary is provided in the [Releases](../../releases) section for
Windows, Linux and macOS — just download and drop it in the game folder.

If you prefer, you can **download the source and build it yourself**:

```bash
# requires Go 1.25+
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

El progreso se muestra en una pequeña **ventana nativa** con **tiempo estimado
restante** (si no hay entorno gráfico, cae a consola). En Windows lleva el icono
de Project Zomboid.

**No** arranca el juego solo: tras actualizar pregunta **Jugar / Limpiar logs y
jugar / Salir**. El juego se lanza **vía Steam** (`steam://rungameid/108600`), que
abre Steam si está cerrado. "Limpiar logs" borra `%UserProfile%\Zomboid\Logs\` y
los `*console.txt` (nunca tus partidas ni mods).

La carpeta de datos de Zomboid se autodetecta en `%UserProfile%\Zomboid`
(`~/Zomboid`). Puedes cambiarla con la variable `PZL_ZOMBOID_DIR` o con un
`pzlauncher.json` junto al ejecutable: `{ "zomboid_dir": "D:\\otra\\ruta\\Zomboid" }`.

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
