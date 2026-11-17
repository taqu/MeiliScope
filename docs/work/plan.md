# MeiliScope — プロジェクト企画書

**Explore your Meilisearch data.**

| 項目 | 内容 |
|---|---|
| プロジェクト名 | MeiliScope |
| プロジェクト種別 | オープンソース・デスクトップアプリケーション |
| 文書バージョン | 1.0 |
| 作成日 | 2026年11月16日 |
| 開発言語 | Go |
| ライセンス | MIT License |
| 公開先 | GitHub |
| 主要対応環境 | Windows x64 |
| ステータス | 企画・基本設計段階 |

---

## 1. エグゼクティブサマリー

**MeiliScope**は、MeilisearchのIndexやDocumentを手軽に閲覧・探索し、検索クエリの実行や必要に応じたデータ編集を行うための、軽量なスタンドアロン・デスクトップアプリケーションである。

データベースブラウザのような直感的な操作体験を提供し、Meilisearchを使用したアプリケーションの開発、デバッグ、テストを支援する。

本プロジェクトでは、Goを主要開発言語として採用し、Windows環境で追加ランタイムや複雑なインストール作業を必要としないアプリケーションを目指す。

ソースコードはMIT LicenseのもとGitHubで公開し、GitHub Actionsによる継続的インテグレーション、ビルド、リリースパッケージの生成を行う。

MeiliScopeはMeilisearchの包括的な管理コンソールではなく、**IndexやDocumentの閲覧と、検索クエリの検証を中心とした開発者向けツール**として位置づける。

## 2. 背景と課題

Meilisearchは高速な全文検索機能を提供する検索エンジンであり、アプリケーションの検索基盤として利用できる。

一方、Meilisearchを使用した開発では、以下のような作業が頻繁に発生する。

- Indexに登録されているDocumentの確認
- Documentのフィールド構造や格納データの調査
- 検索クエリの動作検証
- Filter、Sort、Facetなどの検索条件の調整
- 想定と異なる検索結果の原因調査
- テスト用Documentの追加・修正・削除
- Indexの統計情報や設定内容の確認

これらの操作はHTTP APIを通じて実行できるが、curl、スクリプト、APIクライアントなどを使用して個別に確認することは煩雑になりやすい。

また、既存のWebベースの管理・閲覧ツールとは異なり、デスクトップアプリケーションとして独立して起動し、ローカル開発環境から手軽に利用できるツールにも価値がある。

MeiliScopeは、このような開発・検証作業を単一のGUIに集約する。

## 3. プロジェクトのミッション

> **Explore your Meilisearch data.**
>
> Meilisearchのデータを、インストールの手間なく、直感的に探索できる環境を提供する。

### 3.1 基本理念

**Simple**

アプリケーションを取得し、実行するだけで利用できることを目指す。

**Lightweight**

不要なバックグラウンドサービスや複雑な依存関係を持たず、軽量に動作する。

**Developer-Friendly**

検索クエリの実行、Documentの確認、APIエラーの調査など、開発者が日常的に必要とする操作を重視する。

**Safe by Default**

閲覧を基本とし、Documentの変更・削除は明示的な操作によってのみ実行する。

**Open Source**

MIT Licenseで公開し、利用、改変、再配布しやすいプロジェクトとする。

## 4. プロジェクトの目的

MeiliScopeは以下の5つを主要目標とする。

### 4.1 Indexの探索

Meilisearchに存在するIndexを一覧し、その構造や統計情報を手軽に確認できるようにする。

### 4.2 Documentの閲覧

Documentをテーブル形式およびJSON形式で閲覧できるようにし、データベースブラウザと同様の操作体験を提供する。

### 4.3 クエリの実行と検証

Meilisearch Search APIに対して任意のJSONクエリを実行し、検索結果、ランキング情報、エラー、処理時間を確認できるようにする。

### 4.4 必要最小限のデータ編集

Documentの追加、修正、削除を可能にする。ただし、編集機能は補助的な位置づけとし、管理機能の拡大を優先しない。

### 4.5 導入・配布の簡素化

Windows環境では、ZIPファイルを展開して実行ファイルを起動するだけで利用できる配布形態を目指す。

## 5. 対象ユーザー

| 対象ユーザー | 主な利用目的 |
|---|---|
| アプリケーション開発者 | IndexやDocumentの確認、検索機能のデバッグ |
| バックエンド開発者 | APIクエリ、Filter、Sort、Facetの動作検証 |
| 検索システム開発者 | 検索結果やランキング情報の調査 |
| QA・テスト担当者 | テストデータや検索結果の検証 |
| ローカル開発環境の利用者 | Meilisearch内のデータの探索・修正 |

主な利用場面は、ローカル開発、テスト、デバッグである。

リモートのMeilisearchにも接続可能とするが、本番環境の運用管理や監視を主要用途とはしない。

## 6. 製品スコープ

### 6.1 基本的な位置づけ

MeiliScopeは、Meilisearch専用のデータブラウザとして開発する。

機能の中心は以下の3領域である。

1. **Explore** — IndexとDocumentの閲覧・探索
2. **Query** — 検索クエリの実行・検証
3. **Edit** — 必要最小限のDocument編集

閲覧とクエリ実行を最優先し、編集機能はその次の優先順位とする。

### 6.2 主要機能と優先度

| 優先度 | 機能 | 内容 |
|---|---|---|
| S | Connection Manager | Meilisearch接続先の登録・切り替え |
| S | Index Explorer | Index一覧と統計情報の閲覧 |
| S | Document Browser | Document一覧、ページング、詳細閲覧 |
| S | JSON Inspector | JSON構造の表示・探索 |
| S | Query Console | 任意のSearch APIクエリの実行 |
| S | Query Results | Table / JSON形式での結果表示 |
| A | Search Builder | GUIによる検索条件の構築 |
| A | Document Editor | Documentの追加・編集・削除 |
| A | Query History | クエリ履歴、保存、再実行 |
| A | Export | JSON / CSV出力 |
| A | Task Tracking | Document更新Taskの結果確認 |
| B | Index Settings Viewer | Index設定の閲覧 |
| B | HTTP Request Console | Search API以外のHTTP API呼び出し |
| 対象外 | Server Administration | サーバー運用管理、API Key発行など |

優先度SはMVPに必須、Aは正式リリースまでの主要機能、Bは開発状況に応じて追加する機能とする。

## 7. 主要機能仕様

### 7.1 Connection Manager

複数のMeilisearchインスタンスを登録・切り替えできる機能を提供する。

主な機能：

- 接続名、URL、API Keyの登録
- 接続テスト
- 接続情報の保存と変更
- 接続先の切り替え
- 接続エラーの表示
- HTTPS接続への対応
- 読み取り専用API Keyでの利用

API Keyなどの認証情報は、可能な限りOSのセキュアな資格情報管理機構を利用する。

認証情報をログ、クエリ履歴、平文設定ファイルへ保存しない。

### 7.2 Index Explorer

接続先のIndexを一覧表示する。

主な表示情報：

- Index UID
- Primary Key
- Document数
- Index作成・更新日時
- Index統計情報
- フィールド構造

Indexごとの情報を参照しながら、Document BrowserやQuery Consoleへ移動できるようにする。

### 7.3 Document Browser

Index内のDocumentをテーブル形式で表示する。

主な機能：

- Document一覧
- ページング
- 表示フィールドの選択
- 列幅の変更
- Documentの個別取得
- JSON Inspectorとの連携
- 大量データへの対応
- JSON / CSVエクスポート

Meilisearchは固定スキーマを前提としないため、Documentごとにフィールドが異なる場合にも対応する。

大量のDocumentを一括取得せず、必要なデータのみ取得・描画する。

### 7.4 JSON Inspector

選択したDocumentの構造を確認できるビューを提供する。

主な機能：

- JSONの整形表示
- オブジェクト・配列の展開と折りたたみ
- ネストしたフィールドの閲覧
- 値のコピー
- JSON全文のコピー
- JSON編集画面への移動

Document Browserと連動し、選択Documentの詳細を表示する。

### 7.5 Query Console

MeiliScopeの中核機能とする。

MeilisearchのSearch APIに対してJSONクエリを直接送信できるエディタを提供する。

クエリ例：

```json
{
  "q": "rendering",
  "filter": "category = 'graphics'",
  "sort": ["created_at:desc"],
  "limit": 20,
  "showRankingScore": true
}
```

主な機能：

- JSONエディタ
- JSON構文チェック
- クエリ実行
- 検索結果のTable / JSON表示
- ヒット件数やページング情報の表示
- Ranking Scoreなどの検索メタデータの表示
- Meilisearchが返す処理時間の表示
- クライアント側の実行経過時間の表示
- APIエラーの詳細表示
- クエリ履歴、保存、再実行

Query ConsoleはMeilisearchが提供する検索パラメータをできるだけ透過的に扱う。

MeilisearchのAPI仕様変更に追従できるよう、UIでサポートするパラメータを固定しすぎない設計とする。

### 7.6 Search Builder

JSONを直接記述しなくても検索条件を指定できる補助機能を提供する。

主な入力項目：

- 検索キーワード
- Filter
- Sort
- Limit / Offset
- Facets
- 表示対象フィールド

Search Builderで構築したクエリをQuery Consoleで確認・編集できるようにする。

Search BuilderはQuery Consoleとは独立した検索機構を持たず、共通のクエリ実行処理を使用する。

### 7.7 Document Editor

Documentを必要に応じて編集できる機能を提供する。

主な機能：

- Document追加
- Document編集
- Document削除
- JSON編集
- 変更差分の確認
- 保存前の確認
- 更新Taskの状態追跡
- 編集結果の再取得

MeilisearchのDocument更新は非同期処理となるため、リクエストの受理と更新の完了を区別する。

編集失敗時にはTaskのエラー内容を表示する。

編集可能な権限がない場合は、編集操作を無効化するか、適切な権限エラーを表示する。

### 7.8 HTTP Request Console

Search API以外のMeilisearch HTTP APIを実行する補助機能を将来的に提供する。

GETなどの読み取り操作を基本とし、データの変更を伴うリクエストについては、送信前に明示的な確認を要求する。

この機能はAPIデバッグを目的とし、包括的なサーバー管理コンソールへ発展させることを目的としない。

## 8. ユーザーインターフェース

### 8.1 設計方針

データベースブラウザや開発者向けツールに近い、情報密度が高く操作しやすいUIを採用する。

基本的な画面構成：

- 上部：接続先選択、接続状態、共通操作
- 左側：Index Explorer
- 中央：Document Browser / Query Console
- 右側または下部：Document Inspector / Query Results
- 下部：ステータス、エラー、Taskの状態

### 8.2 UI設計原則

- 閲覧操作を中心とする。
- 検索とDocument表示を素早く切り替えられる。
- クエリエディタと結果表示を一体化する。
- JSONの構造を把握しやすくする。
- 大量データでも操作が重くならないようにする。
- 破壊的操作は通常の閲覧操作から明確に分離する。
- キーボード操作とショートカットを重視する。
- 不要なダイアログや画面遷移を減らす。

## 9. 技術方針

### 9.1 技術スタック

| 領域 | 採用方針 |
|---|---|
| 主要開発言語 | Go |
| GUIフレームワーク | Wails v3（第一候補、要検証） |
| Frontend | TypeScript + React |
| JSONエディタ | Monaco Editor（候補） |
| テーブル表示 | TanStack Table / Virtual（候補） |
| HTTP通信 | Go標準ライブラリ `net/http` |
| 設定保存 | ローカル設定ファイル |
| 認証情報 | OS資格情報管理機構 |
| ビルド | Go / Wails Toolchain |
| CI/CD | GitHub Actions |
| 配布 | GitHub Releases |

Goを選択する主な理由は、実行時の外部依存を抑えやすく、Windows向けに単一実行ファイルを生成でき、HTTP APIクライアントの実装にも適しているためである。

Wails v3の採用にあたっては、APIの安定性、ビルド方法、WebView2依存、単一EXE化、配布環境での起動を事前に検証する。

### 9.2 アーキテクチャ

基本構造：

```text
MeiliScope Desktop Application
│
├── Frontend (React / TypeScript)
│   ├── Connection Manager UI
│   ├── Index Explorer
│   ├── Document Browser
│   ├── JSON Inspector
│   ├── Query Console
│   └── Document Editor
│
├── Application Backend (Go)
│   ├── Connection Manager
│   ├── Meilisearch API Client
│   ├── Query Executor
│   ├── Document Service
│   ├── Task Tracker
│   ├── Settings Manager
│   └── Credential Storage
│
└── Meilisearch HTTP API
    ├── Local Server
    ├── LAN Server
    └── Remote HTTPS Server
```

Frontendは表示と操作を担当し、MeilisearchとのHTTP通信はGo Backendに集約する。

通信処理はUIから分離し、タイムアウト、キャンセル、リトライ、エラー処理を統一的に扱う。

### 9.3 データ保存方針

アプリケーションが保持する情報を以下に限定する。

- 接続先設定
- UI設定
- クエリ履歴
- 保存クエリ
- 必要なローカルキャッシュ

MeilisearchのDocumentを独自データベースへ恒久的に複製することは、初期スコープに含めない。

認証情報は通常のアプリケーション設定とは分離する。

## 10. 配布・インストール方針

### 10.1 主要ターゲット

初期リリースはWindows x64を正式な対象とする。

LinuxおよびmacOSは将来的な対応対象とし、Windows向けのユーザー体験を損なわない範囲で検討する。

### 10.2 配布形態

GitHub ReleasesでZIP形式のリリースパッケージを提供する。

例：

```text
MeiliScope_v1.0.0_windows_amd64.zip
├── meiliscope.exe
├── LICENSE
└── README.md
```

利用者はZIPを展開し、`meiliscope.exe`を起動する。

Python、Node.js、Go、.NETランタイム、Dockerなどの別途導入を要求しないことを目標とする。

ただし、Wailsが依存するWindows WebView2など、GUI実行環境の前提条件は明示する。

### 10.3 アップデート方針

初期バージョンでは自動更新機能を実装せず、GitHub Releasesから新しいバージョンを取得する方式を基本とする。

ユーザー設定は実行ファイルと分離し、アプリケーションの更新時にも保持されるようにする。

## 11. OSS・ライセンス方針

### 11.1 ライセンス

**MeiliScopeはMIT Licenseで公開する。**

商用利用、改変、再配布を許可し、利用者が自由に導入できるOSSとして開発する。

リポジトリには正式なMIT Licenseの全文を含む`LICENSE`ファイルを配置する。

### 11.2 GitHubリポジトリ

GitHubのPublic Repositoryとして公開する。

主な管理対象：

- ソースコード
- 技術仕様
- 開発手順
- ビルド・テストスクリプト
- GitHub Actions
- Issue / Pull Request
- リリース履歴
- 変更履歴

### 11.3 バージョニング

Semantic Versioningを採用する。

- `0.x.x`：初期開発・仕様変更の可能性がある段階
- `1.0.0`：正式な初期安定版
- `1.x.x`：互換性を考慮した機能追加と修正

## 12. CI/CD・パッケージ管理

GitHub Actionsを利用して、ビルド、テスト、リリースパッケージ生成を自動化する。

### 12.1 Continuous Integration

PushおよびPull Requestを契機に以下を実行する。

- Goコードのフォーマット・静的解析
- Go単体テスト
- FrontendのLint / Type Check
- Frontendビルド
- アプリケーション全体のビルド
- 可能な範囲での統合テスト

### 12.2 Continuous Delivery

バージョンタグを契機としてリリースWorkflowを実行する。

主な処理：

1. ソースコードのチェックアウト
2. ビルド環境の準備
3. GoおよびFrontendのビルド
4. Windows実行ファイルの生成
5. リリースパッケージの作成
6. SHA-256チェックサムの生成
7. GitHub Releasesへの成果物登録

### 12.3 依存関係管理

Go Modulesおよびnpmのロックファイルを利用し、依存関係の再現性を確保する。

Dependabotを使用して依存ライブラリとGitHub Actionsの更新を管理する。

依存ライブラリのライセンスについても、MIT Licenseでの再配布と矛盾しないことを確認する。

## 13. ドキュメント方針

### 13.1 基本方針

ドキュメントはGitHubリポジトリ内で管理する。

初期段階ではREADMEを入口とし、詳細資料を`docs/`以下に配置する。

### 13.2 ドキュメント構成

```text
docs/
├── en/
│   ├── getting-started/
│   ├── user-guide/
│   ├── development/
│   └── architecture/
│
└── ja/
    ├── getting-started/
    ├── user-guide/
    ├── development/
    └── architecture/
```

主要な資料：

- README
- Installation Guide
- Quick Start
- User Guide
- Query Console Guide
- Developer Guide
- Architecture
- Build Instructions
- Troubleshooting
- CONTRIBUTING
- CHANGELOG

### 13.3 GitHub Wiki

GitHub Wikiは初期段階では必須としない。

ドキュメント量が増加し、リポジトリ内のMarkdownだけでは整理しにくくなった段階で導入を検討する。

ソースコードとバージョンを同期する必要がある仕様書や開発手順は、引き続き`docs/`で管理する。

GitHub Wikiは補足情報やナレッジベースとしての利用を想定する。

## 14. 非機能要件

| 分類 | 要件 |
|---|---|
| 導入性 | ZIP展開のみで利用できることを目指す |
| 起動性 | 別プロセスの専用サーバーを手動起動する必要がない |
| パフォーマンス | 大量DocumentでもUI操作をブロックしない |
| メモリ | 全Documentを一括保持しない |
| 安定性 | 通信エラーや接続切断から復帰できる |
| セキュリティ | API Keyを安全に取り扱う |
| データ保護 | 変更・削除前に明示的な確認を行う |
| 可搬性 | アプリケーション本体と設定を分離する |
| 保守性 | UIとAPI通信処理の責務を分離する |
| テスト | API処理をGUIに依存せず検証できる |
| 互換性 | 対応Meilisearchバージョンを明確にする |
| プライバシー | クラウドサービスへのデータ送信を必須としない |

## 15. 対象外の機能

初期開発では以下の機能を実装しない。

- Meilisearchサーバーのインストール・起動・停止
- Meilisearchのクラスタ管理
- サーバーの運用監視
- API Keyの発行・管理UI
- ユーザー権限管理システム
- Indexの自動同期やバックアップ
- 検索データの独自保存エンジン
- 複数ユーザーによる共同編集
- 独自の全文検索エンジン
- Webサービスとしてのホスティング

MeiliScopeはMeilisearch本体を操作するための軽量なクライアントであり、検索エンジンやサーバー管理システムの代替とはしない。

## 16. 開発ロードマップ

### Phase 0 — Foundation

**目的：開発基盤と配布基盤の確立**

- Go / Wailsプロジェクトの初期化
- 基本的なGUIの起動
- GitHubリポジトリの構成
- MIT Licenseの配置
- GitHub ActionsによるCI構築
- Windows EXEビルド
- ZIPリリースパッケージの生成
- READMEおよび開発手順の整備

**完了条件：** GitHub ActionsからWindows向けアプリケーションをビルドでき、生成物を起動できる。

### Phase 1 — Connection & Index Explorer

**目的：Meilisearchへの接続とIndex閲覧**

- Connection Manager
- API Key認証
- 接続情報保存
- Index一覧
- Index統計情報
- 接続状態・エラー表示

**完了条件：** Meilisearchに接続し、Index一覧と主要情報をGUIで確認できる。

### Phase 2 — Document Browser

**目的：Documentの探索機能を実装**

- Document一覧
- テーブル表示
- ページング
- JSON Inspector
- Document個別取得
- 可変フィールド構造への対応

**完了条件：** Index内のDocumentをテーブルとJSONで閲覧できる。

### Phase 3 — Query Console

**目的：任意の検索クエリを実行・検証する**

- JSON Query Editor
- Search APIとの連携
- Query Results
- Table / JSON切り替え
- APIエラー表示
- クエリ履歴

**完了条件：** 任意のSearch APIクエリを実行し、レスポンスを確認できる。

**Phase 3をMVPの完成時点とする。**

### Phase 4 — Document Editing

**目的：補助的なDocument編集機能を追加**

- Document追加
- Document編集
- Document削除
- 差分確認
- 更新確認
- Task状態追跡

**完了条件：** GUIからDocumentを安全に変更し、非同期更新の結果を確認できる。

### Phase 5 — Usability & Release

**目的：正式リリースに向けた品質改善**

- Search Builder
- 保存クエリ
- JSON / CSV Export
- キーボードショートカット
- UIの改善
- エラー処理の強化
- ドキュメントの整備
- 統合テスト
- リリースパッケージの検証

**完了条件：** 主要機能が安定して動作し、Windows環境で追加の開発ツールを導入せず利用できる。

### 将来的な拡張

必要性に応じて以下を検討する。

- Linux対応
- macOS対応
- HTTP Request Consoleの拡充
- Index Settingsの閲覧強化
- Query Resultsの比較
- 検索結果の詳細な診断機能

ただし、機能拡張によって軽量性や基本操作の簡潔さを損なわないことを原則とする。

## 17. 技術的リスクと対応方針

### 17.1 GUIフレームワークへの依存

Wails v3の仕様やビルド環境によっては、配布方法や対応プラットフォームに制約が生じる可能性がある。

**対応方針：** Phase 0で実行ファイルの生成、WebView2依存関係、配布先PCでの起動を検証する。

### 17.2 Meilisearch APIのバージョン差異

MeilisearchのバージョンによってAPI仕様や利用可能な検索機能が異なる可能性がある。

**対応方針：** 対応バージョンを定義し、APIクライアントの互換性をテストする。

### 17.3 大量Documentの表示負荷

大量の検索結果やネストしたJSONによって、UI描画やメモリ使用量が増える可能性がある。

**対応方針：** サーバー側ページング、仮想スクロール、必要時の詳細取得を採用する。

### 17.4 意図しないデータ変更

Document編集や削除によって、接続先のデータを意図せず変更する可能性がある。

**対応方針：** 閲覧を既定動作とし、変更内容の確認、Task結果の追跡、適切な権限チェックを実装する。

### 17.5 認証情報の漏えい

API Keyの保存やデバッグログによって認証情報が漏えいする可能性がある。

**対応方針：** OS資格情報管理機構を利用し、ログと履歴への認証情報混入を防ぐ。

## 18. 成功基準

MeiliScopeの初期リリースは、以下の操作を一貫したGUI上で完了できることを成功基準とする。

1. GitHub Releasesからアプリケーションを取得し、ZIPを展開する。
2. EXEを起動してMeilisearchに接続する。
3. 接続先のIndexを一覧表示する。
4. Index内のDocumentをテーブル形式で閲覧する。
5. 選択したDocumentをJSON形式で確認する。
6. 任意の検索クエリを実行する。
7. 検索結果とAPIエラーを確認する。
8. 必要に応じてDocumentを編集する。
9. 編集結果がMeilisearchに反映されたことを確認する。

これらの操作を、Python、Node.js、Dockerなどの追加開発環境を要求せずに行えることを主要な達成指標とする。

## 19. 今後の決定事項

以下は企画段階での未確定事項とし、Technical ProposalおよびPhase 0で決定する。

| 項目 | 現時点の方針 |
|---|---|
| GUIフレームワーク | Wails v3を第一候補として検証 |
| 対応Meilisearchバージョン | 実装・テスト時に決定 |
| クエリエディタ | Monaco Editorを候補とする |
| テーブルコンポーネント | TanStack Table / Virtualを候補とする |
| 設定保存形式 | JSONファイルを基本に検討 |
| リリース署名 | 配布・運用コストを考慮して検討 |
| Linux / macOS | Windows版の安定後に検討 |
| GitHub Wiki | ドキュメント量に応じて導入 |

## 20. 結論

**MeiliScopeは、Meilisearchのデータ探索とクエリ検証に特化した、軽量なオープンソース・デスクトップアプリケーションとして開発する。**

開発言語にはGoを採用し、Windows向けスタンドアロンアプリケーションを最初のターゲットとする。

製品の中心はIndex Explorer、Document Browser、Query Consoleの3機能とし、Document編集は補助機能として提供する。

プロジェクトはMIT LicenseでGitHubに公開し、GitHub ActionsによるCI/CD、GitHub Releasesによるパッケージ配布を行う。ドキュメントはリポジトリ内で管理し、必要に応じてGitHub Wikiを導入する。

機能を際限なく増やすのではなく、**「Meilisearchの中身を確認したいとき、すぐに起動して使えるツール」**という一貫した価値を維持する。

---

**MeiliScope**

*Explore your Meilisearch data.*

*End of Project Proposal — Version 1.0*
