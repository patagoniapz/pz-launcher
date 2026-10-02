# payload/

Deja aquí los ficheros del juego que quieras distribuir en el próximo release.

El caso habitual es un único fichero:

```
payload/projectzomboid.jar
```

Al publicar una etiqueta `vX.Y.Z`, la CI copia el contenido de esta carpeta,
calcula su SHA-256 y lo incluye en el `manifest.json`. El launcher comparará ese
hash con la copia local de cada usuario y descargará solo lo que haya cambiado.

> Para gestionar más ficheros, añádelos aquí y amplía el paso
> "Generar manifest.json" del workflow con más entradas `-game ruta=dist/fichero`.
