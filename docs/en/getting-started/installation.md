# Installation

## Windows

### System Requirements

- Windows 10 version 1803 (Build 17134) or later, or Windows 11 (all versions)
- Microsoft Edge WebView2 Runtime

#### About WebView2

WebView2 is the browser engine used by MeiliScope to render its interface. It is included with Windows 10 version 1803 and later and all versions of Windows 11, and is kept up to date automatically via Windows Update.

If WebView2 is absent (older Windows 10 or a stripped installation), MeiliScope will attempt a silent automatic download on first launch. An internet connection is required for this automatic installation.

You do **not** need Go, Node.js, Python, .NET, or any other developer tool.

### Installation Steps

1. Download `MeiliScope_<version>_windows_amd64.zip` from [GitHub Releases](https://github.com/taqu/meiliscope/releases).
2. Verify the SHA-256 checksum (recommended):
   ```powershell
   Get-FileHash MeiliScope_<version>_windows_amd64.zip -Algorithm SHA256
   ```
   Compare the output against the value in `SHA256SUMS.txt`.
3. Extract the ZIP to any folder (e.g. `C:\Tools\MeiliScope\`).
4. Run `meiliscope.exe`.

No administrator privileges are required.

### First Launch

On first launch MeiliScope creates its configuration directory at:

```
%AppData%\MeiliScope\
```

Typically `C:\Users\<YourName>\AppData\Roaming\MeiliScope\`. Configuration persists across updates.

### Updating

1. Download and extract the new ZIP to the same folder, replacing `meiliscope.exe`.
2. Configuration in `%AppData%\MeiliScope\` is preserved automatically.

### Uninstalling

1. Delete `meiliscope.exe` and any other files from the installation folder.
2. Optionally delete `%AppData%\MeiliScope\` to remove configuration and logs.

No registry entries are created. No system-wide changes are made.

### Windows Troubleshooting

**The application does not start**
- Confirm Windows version is 10 1803 or later.
- Confirm WebView2 is installed: **Settings → Apps → Installed apps**, search for "WebView2".
- Check the log file at `%AppData%\MeiliScope\logs\` for error messages.

**A blank white window appears**
WebView2 initialisation issue. Update WebView2 via Windows Update.

**Windows Defender SmartScreen warning**
MeiliScope is not code-signed. Click **More info → Run anyway**. This warning is expected for unsigned applications downloaded from the internet.

---

## Linux

### System Requirements

MeiliScope requires a 64-bit (amd64) Linux system with a graphical desktop session.

Validated on **Ubuntu 22.04 LTS**. Other distributions with equivalent library versions should work but are not formally tested.

Required runtime libraries:

| Library | Package (Debian/Ubuntu) |
|---|---|
| GTK 3 | `libgtk-3-0` |
| WebKit2GTK 4.0 | `libwebkit2gtk-4.0-37` |

Install them:
```bash
sudo apt-get install libgtk-3-0 libwebkit2gtk-4.0-37
```

> **Note:** You do **not** need Go, Node.js, or any other development tool to run MeiliScope.

### Installation Steps

1. Download `MeiliScope_<version>_linux_amd64.tar.gz` from [GitHub Releases](https://github.com/taqu/meiliscope/releases).
2. Verify the SHA-256 checksum:
   ```bash
   sha256sum -c SHA256SUMS.txt --ignore-missing
   ```
3. Extract the archive:
   ```bash
   tar -xzf MeiliScope_<version>_linux_amd64.tar.gz
   ```
4. Ensure the binary is executable:
   ```bash
   chmod +x meiliscope
   ```
5. Run:
   ```bash
   ./meiliscope
   ```

### Credential Storage on Linux

MeiliScope stores API keys using the **Secret Service** D-Bus API (compatible with GNOME Keyring and KWallet).

If the Secret Service daemon is unavailable (headless environments, locked sessions), MeiliScope falls back to **session-only** mode: credentials are kept in process memory and are lost on exit. The connection panel displays a warning when session mode is active.

### Configuration Location

```
~/.local/share/MeiliScope/
```

or the path given by `$XDG_DATA_HOME/MeiliScope/` if `XDG_DATA_HOME` is set.

### Linux Troubleshooting

**The application does not start**
- Confirm required libraries are installed: `ldconfig -p | grep webkit2gtk`
- Confirm you have a display session (`echo $DISPLAY` or `echo $WAYLAND_DISPLAY`)
- Check `~/.local/share/MeiliScope/logs/` for error messages

**WebKit2GTK version mismatch**
MeiliScope requires `libwebkit2gtk-4.0`. On newer systems where only `libwebkit2gtk-4.1` or `libwebkitgtk-6.0` is available, run from a compatible distribution or container.

**Credentials lost on restart**
The Secret Service daemon is not running. Ensure GNOME Keyring or KWallet is active, or accept session-only mode.
