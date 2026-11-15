# Installation

## System requirements

- **Windows 10** version 1803 (Build 17134) or later
- **Windows 11** (all versions)
- Microsoft Edge WebView2 Runtime

### About WebView2

WebView2 is the browser engine used by MeiliScope to render its interface. It is included with Windows 10 version 1803 and later, and with all versions of Windows 11. It is kept up to date automatically via Windows Update.

If WebView2 is not present on your machine (older Windows 10, or a stripped installation), MeiliScope will attempt to download and install it silently on first launch. An internet connection is required for this automatic installation.

You do **not** need to install Go, Node.js, Python, .NET, or any other developer tools.

---

## Installation

MeiliScope is distributed as a portable ZIP file. No installer is required.

1. Download `MeiliScope_<version>_windows_amd64.zip` from [GitHub Releases](https://github.com/taqu/meiliscope/releases).
2. Verify the SHA-256 checksum (optional but recommended):
   ```powershell
   Get-FileHash MeiliScope_<version>_windows_amd64.zip -Algorithm SHA256
   ```
   Compare the output to the contents of `MeiliScope_<version>_windows_amd64.zip.sha256`.
3. Extract the ZIP to any folder (e.g., `C:\Tools\MeiliScope\`).
4. Run `meiliscope.exe`.

No administrator privileges are required.

---

## First launch

On first launch, MeiliScope creates its configuration directory at:

```
%AppData%\MeiliScope\
```

This is typically `C:\Users\<YourName>\AppData\Roaming\MeiliScope\`.

Configuration persists across application updates. To reset to defaults, delete this directory.

---

## Updating

To update MeiliScope:

1. Download and extract the new ZIP to the same folder, replacing `meiliscope.exe`.
2. Your configuration in `%AppData%\MeiliScope\` is preserved.

---

## Uninstallation

1. Delete `meiliscope.exe` and any other files from the installation folder.
2. Optionally delete `%AppData%\MeiliScope\` to remove configuration and logs.

No registry entries are created. No system-wide changes are made.

---

## Troubleshooting

### The application does not start

- Verify that your Windows version is 10 1803 or later.
- Check that WebView2 is installed: open **Settings → Apps → Installed apps** and search for "WebView2".
- Check the log file at `%AppData%\MeiliScope\logs\` for error messages.

### A blank white window appears

This typically indicates a WebView2 initialisation issue. Ensure WebView2 is up to date via Windows Update.

### Windows Defender SmartScreen warning

Because MeiliScope is not yet code-signed, Windows may show a SmartScreen warning on first launch. Click **More info → Run anyway**. Code signing is planned before the v1.0.0 release.
