package main

import (
	"io/fs"
	"log"
	"net/http"

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
	seed.Load(repo)

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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("POST /api/auth/register", authH.Register)
	mux.HandleFunc("POST /api/auth/login", authH.Login)
	mux.Handle("GET /api/player/me", middleware.RequireAuth(repo, http.HandlerFunc(playerH.Me)))
	mux.Handle("GET /api/areas", middleware.RequireAuth(repo, http.HandlerFunc(areaH.List)))
	mux.Handle("POST /api/explore", middleware.RequireAuth(repo, http.HandlerFunc(exploreH.Explore)))
	mux.Handle("GET /api/items/inventory", middleware.RequireAuth(repo, http.HandlerFunc(itemH.Inventory)))
	mux.Handle("GET /api/encyclopedia", middleware.RequireAuth(repo, http.HandlerFunc(itemH.Encyclopedia)))
	mux.Handle("POST /api/login-bonus/claim", middleware.RequireAuth(repo, http.HandlerFunc(loginBonusH.Claim)))
	mux.Handle("GET /api/notices", middleware.RequireAuth(repo, http.HandlerFunc(noticeH.List)))
	mux.Handle("GET /api/cards/me", middleware.RequireAuth(repo, http.HandlerFunc(cardH.Me)))
	mux.Handle("GET /api/cards/deck", middleware.RequireAuth(repo, http.HandlerFunc(cardH.Deck)))
	mux.Handle("GET /api/cards/collection", middleware.RequireAuth(repo, http.HandlerFunc(cardH.Collection)))
	mux.Handle("POST /api/cards/upgrade", middleware.RequireAuth(repo, http.HandlerFunc(cardH.Upgrade)))
	mux.Handle("POST /api/cards/deck", middleware.RequireAuth(repo, http.HandlerFunc(cardH.UpdateDeck)))
	mux.Handle("POST /api/gacha/draw", middleware.RequireAuth(repo, http.HandlerFunc(cardH.Gacha)))
	mux.Handle("GET /api/boss", middleware.RequireAuth(repo, http.HandlerFunc(bossH.GetBoss)))
	mux.Handle("POST /api/boss/auto", middleware.RequireAuth(repo, http.HandlerFunc(bossH.AutoBattle)))
	mux.Handle("GET /api/checkpoints/master", middleware.RequireAuth(repo, http.HandlerFunc(checkpointH.Master)))
	mux.Handle("GET /api/checkpoints/history", middleware.RequireAuth(repo, http.HandlerFunc(checkpointH.History)))
	mux.Handle("POST /api/checkpoints/claim", middleware.RequireAuth(repo, http.HandlerFunc(checkpointH.Claim)))

	staticFS, err := fs.Sub(webassets.Assets, "static")
	if err != nil {
		log.Fatalf("failed to load web assets: %v", err)
	}
	fileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			http.Error(w, "index not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})

	addr := ":" + cfg.Port
	log.Printf("%s started on %s", cfg.AppName, addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
