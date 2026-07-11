package main

import (
	"bytes"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"retro-treasure-backend/internal/config"
	"retro-treasure-backend/internal/handler"
	"retro-treasure-backend/internal/middleware"
	"retro-treasure-backend/internal/repository"
	"retro-treasure-backend/internal/seed"
	"retro-treasure-backend/internal/service"
	"retro-treasure-backend/internal/webassets"
)

func main() {
	cfg := config.Load()
	repo := repository.NewMemoryRepository()
	repo.SetPersistencePath(cfg.PersistencePath)
	seed.Load(repo)
	if err := repo.LoadPersistentState(); err != nil {
		log.Fatalf("failed to load persistent state: %v", err)
	}

	authSvc := service.NewAuthService(repo)
	playerSvc := service.NewPlayerService(repo)
	areaSvc := service.NewAreaService(repo)
	exploreSvc := service.NewExploreService(repo)
	itemSvc := service.NewItemService(repo)
	loginBonusSvc := service.NewLoginBonusService(repo)
	noticeSvc := service.NewNoticeService(repo)
	cardSvc := service.NewCardService(repo)
	bossSvc := service.NewBossService(repo)
	checkpointSvc := service.NewCheckpointService(repo)

	authH := handler.NewAuthHandler(authSvc)
	playerH := handler.NewPlayerHandler(playerSvc)
	areaH := handler.NewAreaHandler(areaSvc)
	exploreH := handler.NewExploreHandler(exploreSvc)
	itemH := handler.NewItemHandler(itemSvc)
	loginBonusH := handler.NewLoginBonusHandler(loginBonusSvc)
	noticeH := handler.NewNoticeHandler(noticeSvc)
	cardH := handler.NewCardHandler(cardSvc)
	bossH := handler.NewBossHandler(bossSvc)
	checkpointH := handler.NewCheckpointHandler(checkpointSvc)

	appMux := http.NewServeMux()
	appMux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	appMux.HandleFunc("POST /api/auth/register", authH.Register)
	appMux.HandleFunc("POST /api/auth/login", authH.Login)
	appMux.Handle("GET /api/player/me", middleware.RequireAuth(repo, http.HandlerFunc(playerH.Me)))
	appMux.Handle("GET /api/areas", middleware.RequireAuth(repo, http.HandlerFunc(areaH.List)))
	appMux.Handle("POST /api/explore", middleware.RequireAuth(repo, http.HandlerFunc(exploreH.Explore)))
	appMux.Handle("GET /api/items/inventory", middleware.RequireAuth(repo, http.HandlerFunc(itemH.Inventory)))
	appMux.Handle("GET /api/encyclopedia", middleware.RequireAuth(repo, http.HandlerFunc(itemH.Encyclopedia)))
	appMux.Handle("POST /api/login-bonus/claim", middleware.RequireAuth(repo, http.HandlerFunc(loginBonusH.Claim)))
	appMux.Handle("GET /api/notices", middleware.RequireAuth(repo, http.HandlerFunc(noticeH.List)))
	appMux.Handle("GET /api/cards/me", middleware.RequireAuth(repo, http.HandlerFunc(cardH.Me)))
	appMux.Handle("GET /api/cards/deck", middleware.RequireAuth(repo, http.HandlerFunc(cardH.Deck)))
	appMux.Handle("GET /api/cards/collection", middleware.RequireAuth(repo, http.HandlerFunc(cardH.Collection)))
	appMux.Handle("GET /api/cards/archive", middleware.RequireAuth(repo, http.HandlerFunc(cardH.Archive)))
	appMux.Handle("POST /api/cards/upgrade", middleware.RequireAuth(repo, http.HandlerFunc(cardH.Upgrade)))
	appMux.Handle("POST /api/cards/deck", middleware.RequireAuth(repo, http.HandlerFunc(cardH.UpdateDeck)))
	appMux.Handle("POST /api/gacha/draw", middleware.RequireAuth(repo, http.HandlerFunc(cardH.Gacha)))
	appMux.Handle("GET /api/boss", middleware.RequireAuth(repo, http.HandlerFunc(bossH.GetBoss)))
	appMux.Handle("POST /api/boss/auto", middleware.RequireAuth(repo, http.HandlerFunc(bossH.AutoBattle)))
	appMux.HandleFunc("GET /api/checkpoints/master", checkpointH.Master)
	appMux.Handle("GET /api/checkpoints/history", middleware.RequireAuth(repo, http.HandlerFunc(checkpointH.History)))
	appMux.Handle("POST /api/checkpoints/claim", middleware.RequireAuth(repo, http.HandlerFunc(checkpointH.Claim)))

	staticFS, err := fs.Sub(webassets.Assets, "static")
	if err != nil {
		log.Fatalf("failed to load web assets: %v", err)
	}
	appMux.Handle("GET /static/", http.StripPrefix("/static/", basePathAssetHandler(staticFS, cfg.BasePath)))
	appMux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := readBasePathAsset(staticFS, "index.html", cfg.BasePath)
		if err != nil {
			http.Error(w, "index not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})

	mux := basePathMux(appMux, cfg.BasePath)
	addr := cfg.Addr()
	log.Printf("%s started on %s base_path=%q persistence=%q", cfg.AppName, addr, cfg.BasePath, cfg.PersistencePath)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func basePathMux(app http.Handler, basePath string) http.Handler {
	if basePath == "" {
		return app
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		app.ServeHTTP(w, r)
	})
	mux.Handle(basePath+"/", http.StripPrefix(basePath, app))
	mux.HandleFunc(basePath, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, basePath+"/", http.StatusPermanentRedirect)
	})
	mux.Handle("/", app)
	return mux
}

func basePathAssetHandler(staticFS fs.FS, basePath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		data, err := readBasePathAsset(staticFS, name, basePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if contentType := mime.TypeByExtension(path.Ext(name)); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		http.ServeContent(w, r, path.Base(name), time.Time{}, bytes.NewReader(data))
	})
}

func readBasePathAsset(staticFS fs.FS, name string, basePath string) ([]byte, error) {
	data, err := fs.ReadFile(staticFS, name)
	if err != nil {
		return nil, err
	}
	return rewriteBasePathAsset(name, data, basePath), nil
}

func rewriteBasePathAsset(name string, data []byte, basePath string) []byte {
	if basePath == "" {
		return data
	}

	ext := path.Ext(name)
	if ext != ".html" && ext != ".js" && ext != ".css" {
		return data
	}

	out := string(data)
	out = strings.ReplaceAll(out, "<head>", `<head><meta name="app-base-path" content="`+basePath+`"><script>window.__APP_BASE_PATH__="`+basePath+`";</script>`)
	for _, prefix := range []string{`"`, `'`, "`"} {
		out = strings.ReplaceAll(out, prefix+"/static/", prefix+basePath+"/static/")
		out = strings.ReplaceAll(out, prefix+"/api/", prefix+basePath+"/api/")
	}
	out = strings.ReplaceAll(out, "location.href = \"/static/", "location.href = \""+basePath+"/static/")
	out = strings.ReplaceAll(out, "window.location.href = \"/static/", "window.location.href = \""+basePath+"/static/")
	return []byte(out)
}
