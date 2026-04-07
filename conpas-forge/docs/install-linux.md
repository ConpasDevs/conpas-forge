# Guía de Instalación — Linux

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

**Ubuntu / Debian:**
```bash
curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
sudo apt-get install -y nodejs
```

**Fedora / RHEL / CentOS:**
```bash
curl -fsSL https://rpm.nodesource.com/setup_lts.x | sudo bash -
sudo dnf install -y nodejs
```

**Arch Linux:**
```bash
sudo pacman -S nodejs npm
```

Verifica la instalación:
```bash
node --version
```
Debe mostrar algo como `v20.x.x`.

### 2. Claude Code CLI

Claude Code es la interfaz de línea de comandos de Anthropic para trabajar con IA en tu terminal y editor.

```bash
npm install -g @anthropic-ai/claude-code
```

Verifica:
```bash
claude --version
```

> Si no tienes cuenta de Anthropic, créala en [claude.ai](https://claude.ai) antes de continuar.

### 3. Conexión a internet

El instalador descarga el binario de **Engram** desde GitHub durante la instalación. Necesitas acceso a internet.

---

## Instalación

### Paso 1 — Obtener el instalador

**Opción A — Descargar desde GitHub:**

```bash
curl -L https://github.com/ConpasDevs/conpas-forge/raw/main/dist/conpas-forge-linux-amd64 -o conpas-forge
```

**Opción B — Si un compañero te lo compartió:**

Copia el archivo `conpas-forge-linux-amd64` a una carpeta de tu elección.

### Paso 2 — Dar permisos de ejecución

```bash
chmod +x conpas-forge-linux-amd64
```

### Paso 3 — Ejecutar el instalador

```bash
./conpas-forge-linux-amd64 install
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
WARNING: /home/tuusuario/.conpas-forge/bin is not in your PATH.
```

Agrega la siguiente línea al final de tu archivo de configuración de shell:

**Bash** (`~/.bashrc`):
```bash
echo 'export PATH="$HOME/.conpas-forge/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**Zsh** (`~/.zshrc`):
```bash
echo 'export PATH="$HOME/.conpas-forge/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

---

## Verificar la instalación

Abre una nueva terminal y ejecuta:

```bash
engram --version
```

Si muestra un número de versión, Engram está correctamente instalado y en el PATH.

Para confirmar que los skills están desplegados:

```bash
ls ~/.claude/skills
```

Debes ver una lista de carpetas (sdd-init, sdd-apply, zoho-deluge, etc.).

---

## Solución de problemas frecuentes

| Problema | Solución |
|----------|----------|
| `Permission denied` al ejecutar | Asegúrate de haber ejecutado `chmod +x` sobre el binario |
| El módulo Engram falla al descargar | Verifica tu conexión a internet. Si hay un proxy corporativo, puede estar bloqueando la descarga |
| `claude --version` no funciona | Ejecuta `source ~/.bashrc` (o `~/.zshrc`) y vuelve a intentarlo |
| La interfaz no se muestra bien | Usa un emulador de terminal moderno (GNOME Terminal, Alacritty, Kitty). Evita terminales muy básicos |
| `npm install -g` falla por permisos | Configura npm para instalar paquetes globales sin sudo: [guía npm](https://docs.npmjs.com/resolving-eacces-permissions-errors-when-installing-packages-globally) |

---

## Desinstalación

Para eliminar todo lo instalado por conpas-forge:

```bash
rm -rf ~/.conpas-forge/          # binario de Engram
rm -rf ~/.claude/skills/          # skills desplegados
rm -f  ~/.claude/CLAUDE.md        # configuración de personalidad
```

Para limpiar la entrada de Engram en `settings.json`, edita `~/.claude/settings.json` y elimina el bloque `"engram"` dentro de `"mcpServers"`.
