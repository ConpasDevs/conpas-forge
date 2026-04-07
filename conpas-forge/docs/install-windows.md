# Guía de Instalación — Windows

## ¿Qué instala este programa?

`conpas-forge` es un instalador que configura tu entorno de trabajo con IA. Una vez ejecutado, tendrás:

- **Engram** — memoria persistente para tu asistente de IA (recuerda decisiones y contexto entre sesiones)
- **Skills de SDD** — conjunto de habilidades para que la IA te ayude a planificar y escribir código de forma estructurada
- **Skill de Zoho Deluge** — asistencia especializada para scripting en Zoho
- **CLAUDE.md** — archivo que define el comportamiento y personalidad de tu asistente

---

## Requisitos previos

### 1. Claude Code

Claude Code es la interfaz de Anthropic para trabajar con IA en tu terminal y editor. Es el requisito principal — `conpas-forge` configura el entorno dentro de Claude Code.

Tienes tres formas de instalarlo. **Elige una:**

---

#### Opción A — App de escritorio (recomendada para la mayoría)

La forma más sencilla. No requiere Node.js ni ninguna otra dependencia.

1. Descarga el instalador:
   - [Windows x64 (Intel/AMD)](https://claude.ai/api/desktop/win32/x64/setup/latest/redirect)
   - [Windows ARM64 (Surface Pro X, Copilot+ PCs)](https://claude.ai/api/desktop/win32/arm64/setup/latest/redirect)
2. Ejecuta el `.exe` descargado y sigue los pasos
3. Al finalizar, tendrás la app de Claude Code en tu escritorio

> Esta opción instala una interfaz gráfica. Para usar el comando `claude` desde la terminal (que es lo que usa `conpas-forge`), también debes instalar el CLI con la Opción B o C.

---

#### Opción B — Script de instalación (recomendada si usas la terminal)

Instala el CLI de Claude Code directamente. No requiere Node.js. Se actualiza automáticamente.

1. Abre **PowerShell** y ejecuta:
   ```powershell
   irm https://claude.ai/install.ps1 | iex
   ```
2. Verifica que se instaló correctamente:
   ```powershell
   claude --version
   ```

---

#### Opción C — WinGet

Si ya usas el gestor de paquetes de Windows:

```powershell
winget install Anthropic.ClaudeCode
```

---

### 2. Cuenta de Anthropic

Necesitas una cuenta activa en [claude.ai](https://claude.ai) para autenticarte con Claude Code. Si no tienes, créala antes de continuar.

### 3. Conexión a internet

El instalador descarga el binario de **Engram** desde GitHub durante la instalación.

---

## Instalación

### Paso 1 — Obtener el instalador

Descarga `conpas-forge-windows-amd64.exe` desde la página de releases:

```
https://github.com/ConpasDevs/conpas-forge/releases/latest
```

Haz clic en `conpas-forge-windows-amd64.exe` para descargarlo. Guárdalo en una carpeta de tu elección (por ejemplo `C:\Tools\`).

### Paso 2 — Abrir PowerShell en la carpeta del instalador

1. Navega hasta la carpeta donde guardaste el `.exe`
2. Haz clic derecho mientras mantienes `Shift` → **"Abrir ventana de PowerShell aquí"**

   (O abre PowerShell y navega con `cd C:\ruta\donde\guardaste\`)

### Paso 3 — Ejecutar el instalador

```powershell
.\conpas-forge-windows-amd64.exe install
```

Se abrirá una interfaz interactiva en la terminal.

### Paso 4 — Seguir el asistente de instalación

El instalador te guiará por cinco pantallas:

| Pantalla | Qué hacer |
|----------|-----------|
| **Módulos** | Selecciona qué instalar. Usa `Espacio` para marcar/desmarcar, `Enter` para confirmar |
| **Persona** | Elige el estilo de comunicación de tu IA (ej. "Argentino", "Yoda") |
| **Modelos** | Asigna qué modelo de IA usar para cada rol (puedes dejarlo en los valores por defecto) |
| **Instalando** | El instalador trabaja automáticamente. Espera a que termine |
| **Resumen** | Muestra qué se instaló correctamente y qué advertencias hay |

> Usa las teclas de flecha `↑` `↓` para moverte, `Espacio` para seleccionar, `Enter` para confirmar.

### Paso 5 — Agregar Engram al PATH (si el instalador lo indica)

Si al finalizar ves un aviso como:

```
WARNING: C:\Users\TuUsuario\.conpas-forge\bin is not in your PATH.
```

Ejecuta esto en PowerShell para añadirlo de forma permanente:

```powershell
[Environment]::SetEnvironmentVariable('PATH', $env:PATH + ';C:\Users\TuUsuario\.conpas-forge\bin', 'User')
```

Luego **cierra y vuelve a abrir** PowerShell para que surta efecto.

---

## Verificar la instalación

Abre una nueva ventana de PowerShell y ejecuta:

```powershell
engram --version
```

Si muestra un número de versión, Engram está correctamente instalado y en el PATH.

Para confirmar que los skills están desplegados:

```powershell
ls $env:USERPROFILE\.claude\skills
```

Debes ver una lista de carpetas (sdd-init, sdd-apply, zoho-deluge, etc.).

---

## Solución de problemas frecuentes

| Problema | Solución |
|----------|----------|
| `'conpas-forge' no se reconoce como comando` | Asegúrate de estar en la carpeta donde está el `.exe` y usar `.\` antes del nombre |
| El módulo Engram falla al descargar | Verifica tu conexión a internet. Si hay un proxy corporativo, puede estar bloqueando la descarga |
<<<<<<< HEAD
| `claude --version` no funciona | Cierra y vuelve a abrir PowerShell tras instalar Claude Code |
| La terminal muestra caracteres extraños | Usa **Windows Terminal** en lugar de CMD clásico |

---

## Desinstalación

Para eliminar todo lo instalado por conpas-forge:

1. Borra la carpeta `~/.conpas-forge/` (contiene el binario de Engram)
2. Borra la carpeta `~/.claude/skills/` (contiene los skills)
3. Borra `~/.claude/CLAUDE.md`
4. Edita `~/.claude/settings.json` y elimina la entrada `mcpServers.engram`

> `~` equivale a `C:\Users\TuNombreDeUsuario\`
