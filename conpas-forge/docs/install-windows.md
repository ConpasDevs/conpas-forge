# Guía de Instalación — Windows

## ¿Qué instala este programa?

`conpas-forge` es un instalador que configura tu entorno de trabajo con IA. Una vez ejecutado, tendrás:

- **Engram** — memoria persistente para tu asistente de IA (recuerda decisiones y contexto entre sesiones)
- **Skills de SDD** — conjunto de habilidades para que la IA te ayude a planificar y escribir código de forma estructurada
- **Skill de Zoho Deluge** — asistencia especializada para scripting en Zoho
- **CLAUDE.md** — archivo que define el comportamiento y personalidad de tu asistente

---

## Requisitos previos

Antes de ejecutar el instalador, necesitas tener lo siguiente instalado y funcionando.

### 1. Node.js (versión 18 o superior)

Claude Code requiere Node.js para funcionar.

1. Abre el navegador y ve a [nodejs.org](https://nodejs.org)
2. Descarga la versión **LTS** (la recomendada)
3. Ejecuta el instalador y sigue los pasos (opciones por defecto son correctas)
4. Verifica la instalación abriendo PowerShell y ejecutando:
   ```powershell
   node --version
   ```
   Debe mostrar algo como `v20.x.x`.

### 2. Claude Code CLI

Claude Code es la interfaz de línea de comandos de Anthropic para trabajar con IA en tu terminal y editor.

1. Abre **PowerShell** como administrador
2. Ejecuta:
   ```powershell
   npm install -g @anthropic-ai/claude-code
   ```
3. Verifica:
   ```powershell
   claude --version
   ```

> Si no tienes cuenta de Anthropic, créala en [claude.ai](https://claude.ai) antes de continuar.

### 3. Conexión a internet

El instalador descarga el binario de **Engram** desde GitHub durante la instalación. Necesitas acceso a internet.

---

## Instalación

### Paso 1 — Obtener el instalador

Descarga el archivo `conpas-forge-windows-amd64.exe` desde el repositorio:

```
https://github.com/ConpasDevs/conpas-forge
```

Ve a la carpeta `dist/` del repositorio y descarga el `.exe`.

O bien, si te lo ha compartido un compañero, copia el archivo a una carpeta de tu elección (por ejemplo `C:\Tools\`).

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
| `claude --version` no funciona | Cierra y vuelve a abrir PowerShell tras instalar Node.js y Claude Code |
| La terminal muestra caracteres extraños | Usa **Windows Terminal** en lugar de CMD clásico |

---

## Desinstalación

Para eliminar todo lo instalado por conpas-forge:

1. Borra la carpeta `~/.conpas-forge/` (contiene el binario de Engram)
2. Borra la carpeta `~/.claude/skills/` (contiene los skills)
3. Borra `~/.claude/CLAUDE.md`
4. Edita `~/.claude/settings.json` y elimina la entrada `mcpServers.engram`

> `~` equivale a `C:\Users\TuNombreDeUsuario\`
