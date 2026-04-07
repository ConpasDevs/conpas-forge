# Guía de Instalación — Linux

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

> **Nota:** No existe app de escritorio de Claude Code para Linux. La instalación es únicamente por línea de comandos.

#### Instalación recomendada — Script oficial (sin Node.js)

```bash
curl -fsSL https://claude.ai/install.sh | bash
```

Una vez instalado, verifica:

```bash
claude --version
```

El script se encarga de todo automáticamente y configura las actualizaciones.

#### Alternativa — Homebrew (si ya lo tienes instalado)

```bash
brew install --cask claude-code
```

---

> **Nota sobre Alpine Linux / musl libc:** Si usas Alpine u otra distro con musl en lugar de glibc, necesitas instalar `libgcc` y `libstdc++` antes de ejecutar el script de instalación.

---

### 2. Cuenta de Anthropic

Necesitas una cuenta activa en [claude.ai](https://claude.ai) para autenticarte con Claude Code. Si no tienes, créala antes de continuar.

### 3. Conexión a internet

El instalador descarga el binario de **Engram** desde GitHub durante la instalación.

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
| Error en Alpine / musl | Instala `libgcc` y `libstdc++` antes de ejecutar el script de Claude Code |

---

## Desinstalación

Para eliminar todo lo instalado por conpas-forge:

```bash
rm -rf ~/.conpas-forge/          # binario de Engram
rm -rf ~/.claude/skills/          # skills desplegados
rm -f  ~/.claude/CLAUDE.md        # configuración de personalidad
```

Para limpiar la entrada de Engram en `settings.json`, edita `~/.claude/settings.json` y elimina el bloque `"engram"` dentro de `"mcpServers"`.
