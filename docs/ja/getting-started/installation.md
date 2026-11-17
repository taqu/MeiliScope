# インストール

## Windows

### システム要件

- **Windows 10** バージョン 1803 (Build 17134) 以降、または **Windows 11**（全バージョン）
- Microsoft Edge WebView2 ランタイム

#### WebView2 について

WebView2 は MeiliScope のインターフェースを描画するために使用するブラウザエンジンです。Windows 10 バージョン 1803 以降および Windows 11 のすべてのバージョンには標準搭載されており、Windows Update により自動的に更新されます。

WebView2 がインストールされていない場合（古い Windows 10 や最小構成のインストールなど）、MeiliScope は初回起動時に自動でサイレントインストールを試みます。自動インストールにはインターネット接続が必要です。

Go、Node.js、Python、.NET などの開発ツールをインストールする必要は**ありません**。

### インストール手順

1. [GitHub Releases](https://github.com/taqu/meiliscope/releases) から `MeiliScope_<バージョン>_windows_amd64.zip` をダウンロードします。
2. SHA-256 チェックサムを確認します（推奨）:
   ```powershell
   Get-FileHash MeiliScope_<バージョン>_windows_amd64.zip -Algorithm SHA256
   ```
   出力を `SHA256SUMS.txt` の内容と比較してください。
3. ZIP を任意のフォルダー（例: `C:\Tools\MeiliScope\`）に展開します。
4. `meiliscope.exe` を実行します。

管理者権限は不要です。

### 初回起動

初回起動時、MeiliScope は以下の場所に設定ディレクトリを作成します:

```
%AppData%\MeiliScope\
```

通常は `C:\Users\<ユーザー名>\AppData\Roaming\MeiliScope\` です。設定はアプリケーションを更新しても保持されます。

### 更新方法

1. 新しい ZIP をダウンロードし、同じフォルダーに展開して `meiliscope.exe` を上書きします。
2. `%AppData%\MeiliScope\` の設定は保持されます。

### アンインストール

1. インストールフォルダーから `meiliscope.exe` およびその他のファイルを削除します。
2. 必要に応じて `%AppData%\MeiliScope\` を削除すると、設定とログも削除されます。

レジストリへの書き込みやシステム全体への変更はありません。

### Windows トラブルシューティング

**アプリケーションが起動しない**
- Windows のバージョンが 10 1803 以降であることを確認してください。
- **設定 → アプリ → インストール済みアプリ** で「WebView2」を検索し、インストールされているか確認します。
- `%AppData%\MeiliScope\logs\` のログファイルでエラーメッセージを確認します。

**空白（白い）ウィンドウが表示される**
WebView2 の初期化に問題がある可能性があります。Windows Update で WebView2 を最新版に更新してください。

**Windows Defender SmartScreen の警告**
MeiliScope はコード署名されていないため、初回起動時に SmartScreen の警告が表示される場合があります。**詳細情報 → 実行** をクリックしてください。

---

## Linux

### システム要件

MeiliScope には 64 ビット（amd64）Linux とグラフィカルデスクトップセッションが必要です。

**Ubuntu 22.04 LTS** で動作確認済みです。同等のライブラリバージョンを持つ他のディストリビューションでも動作する可能性がありますが、公式にはテストされていません。

必要なランタイムライブラリ:

| ライブラリ | パッケージ（Debian/Ubuntu） |
|---|---|
| GTK 3 | `libgtk-3-0` |
| WebKit2GTK 4.0 | `libwebkit2gtk-4.0-37` |

インストール:
```bash
sudo apt-get install libgtk-3-0 libwebkit2gtk-4.0-37
```

> **注意:** Go、Node.js、その他の開発ツールのインストールは不要です。

### インストール手順

1. [GitHub Releases](https://github.com/taqu/meiliscope/releases) から `MeiliScope_<バージョン>_linux_amd64.tar.gz` をダウンロードします。
2. SHA-256 チェックサムを確認します:
   ```bash
   sha256sum -c SHA256SUMS.txt --ignore-missing
   ```
3. アーカイブを展開します:
   ```bash
   tar -xzf MeiliScope_<バージョン>_linux_amd64.tar.gz
   ```
4. 実行権限を確認します:
   ```bash
   chmod +x meiliscope
   ```
5. 起動します:
   ```bash
   ./meiliscope
   ```

### Linux の資格情報ストレージ

MeiliScope は API キーの保存に **Secret Service** D-Bus API（GNOME Keyring や KWallet に対応）を使用します。

Secret Service デーモンが利用できない場合（ヘッドレス環境、画面ロック中など）、MeiliScope は**セッション専用**モードにフォールバックします。資格情報はプロセスのメモリにのみ保持され、終了時に失われます。セッション専用モードが有効な場合、接続パネルに警告が表示されます。

### 設定の場所

```
~/.local/share/MeiliScope/
```

`$XDG_DATA_HOME` が設定されている場合は `$XDG_DATA_HOME/MeiliScope/` が使用されます。

### Linux トラブルシューティング

**「共有オブジェクトファイルを開けません」エラー**
不足しているランタイムライブラリをインストールしてください（上記参照）。

**「ディスプレイに接続できません」エラー**
MeiliScope にはグラフィカルデスクトップセッションが必要です。ヘッドレス環境では動作しません。

**再起動後に資格情報が失われる**
Secret Service デーモンが実行されていません。GNOME Keyring または KWallet を起動するか、セッション専用モードを受け入れてください。
