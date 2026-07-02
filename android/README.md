# RELIC RAID Android

Android Studio で開ける RELIC RAID の Android WebView アプリです。

## 役割

- 既存の Go バックエンドが配信する `/static/` ゲーム画面を WebView で表示します。
- 横画面固定、フルスクリーン、JavaScript と localStorage を有効にしています。
- 開発中は Android Emulator から `http://10.0.2.2:8080/static/` を読み込みます。

## 起動手順

1. バックエンドを起動します。

   ```bash
   go run ./cmd/server
   ```

2. Android Studio で `android/` ディレクトリを開きます。

3. Emulator を起動して `app` を Run します。

## 実機で確認する場合

実機は `10.0.2.2` を使えないため、Mac の LAN IP を指定します。

```bash
cp android/local.properties.example android/local.properties
```

`android/local.properties` に次のように設定します。

```properties
GAME_BASE_URL=http://<your-mac-lan-ip>:8080/static/
```

その後 Android Studio で Sync/Run してください。

## 注意

- 現在のバックエンドはインメモリ実装なので、サーバー再起動でログイン状態や進行状態は消えます。
- 本番化するときは `GAME_BASE_URL` をデプロイ済みHTTPS URLへ変更してください。
