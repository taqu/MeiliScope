# インストール

## システム要件

- **Windows 10** バージョン 1803 (Build 17134) 以降
- **Windows 11**（全バージョン）
- Microsoft Edge WebView2 ランタイム

### WebView2 について

WebView2 は MeiliScope のインターフェースを描画するために使用するブラウザエンジンです。Windows 10 バージョン 1803 以降および Windows 11 のすべてのバージョンには標準搭載されており、Windows Update により自動的に更新されます。

WebView2 がインストールされていない場合（古い Windows 10 や最小構成のインストールなど）、MeiliScope は初回起動時に自動でサイレントインストールを試みます。自動インストールにはインターネット接続が必要です。

Go、Node.js、Python、.NET などの開発ツールをインストールする必要は**ありません**。

---

## インストール手順

MeiliScope はポータブルな ZIP ファイルとして配布されます。インストーラーは不要です。

1. [GitHub Releases](https://github.com/taqu/meiliscope/releases) から `MeiliScope_<バージョン>_windows_amd64.zip` をダウンロードします。
2. SHA-256 チェックサムを確認します（任意ですが推奨）:
   ```powershell
   Get-FileHash MeiliScope_<バージョン>_windows_amd64.zip -Algorithm SHA256
   ```
   出力を `MeiliScope_<バージョン>_windows_amd64.zip.sha256` の内容と比較してください。
3. ZIP を任意のフォルダー（例: `C:\Tools\MeiliScope\`）に展開します。
4. `meiliscope.exe` を実行します。

管理者権限は不要です。

---

## 初回起動

初回起動時、MeiliScope は以下の場所に設定ディレクトリを作成します:

```
%AppData%\MeiliScope\
```

通常は `C:\Users\<ユーザー名>\AppData\Roaming\MeiliScope\` です。

設定はアプリケーションを更新しても保持されます。初期設定に戻すには、このディレクトリを削除してください。

---

## 更新方法

1. 新しい ZIP をダウンロードし、同じフォルダーに展開して `meiliscope.exe` を上書きします。
2. `%AppData%\MeiliScope\` の設定は保持されます。

---

## アンインストール

1. インストールフォルダーから `meiliscope.exe` およびその他のファイルを削除します。
2. 必要に応じて `%AppData%\MeiliScope\` を削除すると、設定とログも削除されます。

レジストリへの書き込みやシステム全体への変更はありません。

---

## トラブルシューティング

### アプリケーションが起動しない

- Windows のバージョンが 10 1803 以降であることを確認してください。
- WebView2 がインストールされているか確認します: **設定 → アプリ → インストール済みアプリ** で「WebView2」を検索します。
- エラーメッセージは `%AppData%\MeiliScope\logs\` のログファイルで確認できます。

### 空白（白い）ウィンドウが表示される

WebView2 の初期化に問題がある可能性があります。Windows Update で WebView2 を最新版に更新してください。

### Windows Defender SmartScreen の警告

MeiliScope はまだコード署名されていないため、初回起動時に SmartScreen の警告が表示される場合があります。**詳細情報 → 実行** をクリックしてください。コード署名は v1.0.0 リリース前に対応予定です。
