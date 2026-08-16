[English](README.md) | [简体中文](README_ZH.md) | [日本語](README_JA.md) | [한국어](README_KO.md) | [Español](README_ES.md)

# YouTube Uploader MCP

Un servidor MCP local para subir, consultar y auditar vídeos, además de completar las acciones posteriores a la subida compatibles con YouTube mediante YouTube Data API v3.

Ofrece a los clientes MCP un flujo vinculado al canal para OAuth, subidas inicialmente privadas, programación, miniaturas, subtítulos, inserción en listas de reproducción y verificación de los datos accesibles mediante la API. El cliente MCP puede preparar los metadatos, pero este servidor solo envía los valores que recibe.

> [!IMPORTANT]
> Este README documenta la rama `main` actual, que registra nueve herramientas. La versión etiquetada más reciente es anterior y los scripts de instalación incluidos todavía apuntan al repositorio original. Compila el proyecto desde el código fuente para utilizar las funciones aquí documentadas.

## Qué hace

- Autentica canales de YouTube localmente con OAuth 2.0 y admite varios canales almacenados en caché.
- Verifica el canal OAuth real antes de las operaciones vinculadas al canal.
- Sube vídeos con valores de título, descripción, etiquetas, categoría, idioma, privacidad, programación, audiencia infantil, contenido sintético y notificaciones a suscriptores.
- Utiliza `private` como valor predeterminado para las nuevas subidas, salvo que se indique explícitamente otro estado válido.
- Añade un vídeo subido a una lista de reproducción, carga una miniatura de menos de 2 MiB e inserta una pista de subtítulos.
- Lee con `get_video` los metadatos del vídeo visibles para el propietario.
- Ejecuta con `audit_video` comprobaciones de solo lectura basadas en expectativas, incluidos los recursos de subtítulos y la pertenencia u orden exactos de una lista de reproducción.
- Actualiza de forma segura las declaraciones compatibles sobre audiencia infantil y contenido sintético con protección frente a conflictos mediante ETag.

## Requisitos

- Go `1.24.4` o posterior.
- Un proyecto de Google Cloud con YouTube Data API v3 habilitada.
- Un archivo JSON de cliente de escritorio de Google OAuth.
- Tu cuenta de Google añadida como usuario de prueba mientras la aplicación OAuth esté en modo de prueba.
- Cuota de la API de YouTube para las operaciones que ejecutes.
- Un cliente MCP compatible con servidores stdio locales.

Consulta [youtube_oauth2_setup.md](youtube_oauth2_setup.md) para ver el proceso de configuración de Google Cloud.

## Compilar desde el código fuente

```bash
git clone https://github.com/ai-martin-lau/youtube-uploader-mcp.git
cd youtube-uploader-mcp
go mod download
go build -trimpath -o youtube-uploader-mcp .
```

Crea directorios privados para la caché de tokens y los registros, y mantén privado el archivo del cliente OAuth:

```bash
mkdir -p /absolute/path/private-state /absolute/path/private-logs
chmod 700 /absolute/path/private-state /absolute/path/private-logs
chmod 600 /absolute/path/client_secret.json
```

## Configurar el cliente MCP

Utiliza rutas absolutas. El formato de configuración puede tener un nombre distinto según el cliente MCP, pero la entrada del servidor es la misma:

```json
{
  "mcpServers": {
    "youtube-uploader-mcp": {
      "command": "/absolute/path/youtube-uploader-mcp",
      "args": [
        "-client_secret_file",
        "/absolute/path/client_secret.json",
        "-working_dir",
        "/absolute/path/private-state"
      ],
      "env": {
        "YOUTUBE_UPLOADER_MCP_LOG_DIR": "/absolute/path/private-logs"
      }
    }
  }
}
```

Reinicia el cliente MCP después de guardar la configuración.

## Cómo usarlo

### 1. Conectar un canal

Pídele lo siguiente a tu cliente MCP:

```text
Inicia la autenticación de YouTube. Utiliza el URI de redirección configurado en mi cliente de Google OAuth.
```

El cliente debe llamar a `authenticate` y devolver una URL de autorización de Google. Completa el flujo de consentimiento de Google y proporciona después a la herramienta local `accesstoken` el parámetro de consulta `code` de un solo uso de la URL de redirección de localhost. Trata ese código como información sensible.

A continuación, verifica el canal almacenado localmente en caché:

```text
Enumera mis canales de YouTube autenticados y muestra el ID exacto de cada canal.
```

### 2. Subir como privado

```text
Sube /absolute/path/night-drive.mp4 al canal UCxxxxxxxx como privado.

Título: Night Drive Practice
Descripción: Un vídeo original de práctica de bajo.
Etiquetas: practice,bass,original
ID de categoría: 10
Idioma: en
Creado para niños: false
Contiene contenido sintético: false
Notificar a los suscriptores: false

No lo publiques ni lo añadas todavía a una lista de reproducción. Devuelve el ID del vídeo de YouTube y los valores enviados realmente.
```

Conserva el ID de vídeo devuelto. Para programar un vídeo, incluye un valor `publish_at` en formato RFC 3339 en la solicitud original de `upload_video`; las subidas programadas se envían como privadas hasta que YouTube las publica.

### 3. Verificar antes de las acciones posteriores a la subida

```text
Lee el vídeo VIDEO_ID del canal UCxxxxxxxx. Confirma el canal propietario, la privacidad, el título, la categoría, el idioma, la declaración de audiencia infantil y la declaración de contenido sintético. No cambies nada.
```

Para realizar una comprobación basada en expectativas:

```text
Audita el vídeo VIDEO_ID del canal UCxxxxxxxx. Espera privacidad private, categoría 10, idioma en, creado para niños false y contenido sintético false. No modifiques el vídeo.
```

### 4. Completar las acciones posteriores a la subida compatibles

```text
Para el vídeo VIDEO_ID del canal UCxxxxxxxx:
- sube la miniatura /absolute/path/thumbnail.jpg
- añádelo a la lista de reproducción PLAYLIST_ID
- sube /absolute/path/captions.srt como subtítulos en inglés

Informa por separado del resultado de cada acción. Si la inserción en la lista de reproducción agota el tiempo de espera, audita la lista real antes de volver a intentarlo.
```

`update_video` puede completarse solo parcialmente. Revisa siempre por separado el resultado de cada acción solicitada.

## Referencia de herramientas

| Herramienta | Finalidad |
| --- | --- |
| `authenticate` | Crea la URL de autorización de Google OAuth. |
| `accesstoken` | Intercambia el código de autorización de un solo uso, verifica el canal real y almacena el token localmente en caché. |
| `channels` | Enumera los canales encontrados en la caché local de tokens. No es una búsqueda de canales de YouTube en tiempo real. |
| `refreshtoken` | Actualiza manualmente el token almacenado en caché de un canal. Las herramientas vinculadas al canal también actualizan automáticamente los tokens próximos a caducar. |
| `upload_video` | Sube un vídeo local y envía los metadatos, declaraciones, estado de privacidad, programación y opción de notificación proporcionados. |
| `get_video` | Lee los metadatos accesibles mediante la API de un vídeo visible para el propietario y verifica que pertenece al canal. |
| `audit_video` | Compara los valores reales accesibles mediante la API con las expectativas proporcionadas por quien realiza la llamada, sin modificar YouTube. |
| `update_video_metadata` | Actualiza `self_declared_made_for_kids` o `contains_synthetic_media`, o ambos, mediante lectura-fusión-escritura y protección ETag. |
| `update_video` | Inserta el vídeo en una lista de reproducción, carga una miniatura o inserta una pista de subtítulos. |

## Valores predeterminados y límites importantes

- El valor predeterminado de `upload_video` es `private`.
- `publish_at` debe estar en formato RFC 3339; los vídeos programados se suben como privados.
- `tags` es una cadena separada por comas.
- El archivo de la miniatura debe tener menos de 2 MiB.
- `made_for_kids`, `contains_synthetic_media` y `notify_subscribers` son valores booleanos opcionales. Un `false` explícito se envía como `false`.
- YouTube no permite volver a leer posteriormente `notify_subscribers`. Una subida correcta confirma que se aceptó la solicitud, pero una auditoría posterior no puede demostrar ese valor individual y lo informa como `unverifiable`.
- La inserción en listas de reproducción no es idempotente. Después de un error o de que se agote el tiempo de espera, consulta o audita la lista de reproducción real antes de volver a intentarlo para evitar duplicados.
- Se admite la inserción de subtítulos; no se admite su eliminación ni sustitución.
- No se admite crear listas de reproducción, retirar elementos, reordenarlos, asignar portadas a las listas ni configurar el idioma de una lista.
- Este servidor no expone actualmente cambios posteriores a la subida del título, descripción, etiquetas, categoría, idioma, privacidad, programación ni de la mayoría de los demás campos.
- Los controles exclusivos de YouTube Studio, como capítulos automáticos, lugares o conceptos automáticos, remezclas de Shorts, ajustes predefinidos de moderación de comentarios, tarjetas, pantallas finales y certificación de subtítulos, no se pueden leer ni cambiar de forma fiable mediante este MCP. `audit_video` informa esas expectativas como `unverifiable` en lugar de afirmar que se han cumplido.

## Seguridad y privacidad

- Utiliza únicamente una aplicación Google OAuth que controles.
- El servidor solicita los ámbitos `youtube.upload`, `youtube.readonly` y el ámbito sensible `youtube.force-ssl`. El último es necesario para subir subtítulos.
- Los tokens de acceso y actualización se almacenan localmente en `<working_dir>/.youtube_uploader_channels_cache` con permisos de archivo restringidos y se ocultan en la salida de las herramientas.
- El código de autorización OAuth de un solo uso pasa por el cliente MCP. El registro actual de solicitudes puede guardarlo, así que mantén privado el directorio de registros, no compartas los registros y elimina los antiguos cuando ya no sean necesarios.
- Nunca confirmes en el repositorio `client_secret.json`, la caché de canales ni los registros de MCP.
- La implementación OAuth actual utiliza un valor state fijo y no utiliza PKCE. Ejecútala únicamente en una máquina local de confianza y evita flujos de autorización simultáneos o no solicitados.

## Origen del proyecto

Este proyecto se basa en [anwerj/youtube-uploader-mcp](https://github.com/anwerj/youtube-uploader-mcp) y continúa disponible bajo la licencia MIT. Este repositorio conserva el aviso de copyright original y amplía el servidor con verificación vinculada al canal, declaraciones explícitas durante la subida, lecturas visibles para el propietario, auditorías de políticas de solo lectura y actualizaciones de metadatos con detección de conflictos.

## Contribuir

Se agradecen los issues y pull requests bien delimitados. Describe el comportamiento de la API de YouTube que cambia, evita incluir refactorizaciones no relacionadas en el parche y añade o actualiza las pruebas cuando cambie el comportamiento.

## Licencia

[MIT](LICENSE)
