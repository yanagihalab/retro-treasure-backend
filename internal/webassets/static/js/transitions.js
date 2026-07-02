(function () {
  const overlay = document.createElement("div");
  overlay.className = "screen-transition";
  overlay.innerHTML = '<div class="transition-core"></div>';
  document.documentElement.classList.add("is-app-loading");

  function buildDynamicStage() {
    document.title = "";
    document.documentElement.classList.add("js-premium-stage", "js-no-page-title", "js-smart-game-ui", "js-landscape-game-ui");

    document.querySelectorAll(".hero h1, body:not(.battle-only) main > .panel > .panel-head h2, body.special-battle .battle-panel .panel-head h2").forEach((title) => {
      title.classList.add("js-hidden-page-title");
      title.setAttribute("aria-hidden", "true");
    });

    if (!document.querySelector(".js-stage-effects")) {
      const effects = document.createElement("div");
      effects.className = "js-stage-effects";
      effects.setAttribute("aria-hidden", "true");
      effects.innerHTML = `
        <div class="js-aura js-aura-primary"></div>
        <div class="js-aura js-aura-secondary"></div>
        <div class="js-aura js-aura-accent"></div>
        <div class="js-scanline"></div>
        <div class="js-vignette"></div>
        <div class="js-corner js-corner-tl"></div>
        <div class="js-corner js-corner-br"></div>
      `;
      document.body.prepend(effects);
    }

    const isBattle = document.body.classList.contains("battle-only");

    if (!document.querySelector(".js-orientation-banner")) {
      const orientation = document.createElement("div");
      orientation.className = "js-orientation-banner";
      orientation.setAttribute("aria-hidden", "true");
      orientation.innerHTML = "<span>横画面推奨</span>";
      document.body.appendChild(orientation);
    }

    if (!isBattle && !document.querySelector(".js-game-hud")) {
      const username = localStorage.getItem("retro_username") || "GUEST";
      const hud = document.createElement("div");
      hud.className = "js-game-hud";
      hud.innerHTML = `
        <div class="js-game-brand">RELIC RAID</div>
        <div class="js-game-status">
          <span class="js-resource">AP ∞</span>
          <span class="js-resource">COIN</span>
          <span class="js-player">${username}</span>
        </div>
      `;
      document.body.appendChild(hud);
    }

    if (!isBattle && !document.querySelector(".js-bottom-dock")) {
      const current = document.body.dataset.page || "home";
      const dockItems = [
        ["home", "/static/index.html", "HOME"],
        ["mypage", "/static/mypage.html", "PLAYER"],
        ["deck", "/static/deck.html", "DECK"],
        ["boss", "/static/boss.html", "BOSS"],
        ["checkpoint", "/static/checkpoint.html", "NODE"],
        ["gacha", "/static/gacha.html", "GACHA"],
      ];
      const dock = document.createElement("nav");
      dock.className = "js-bottom-dock";
      dock.setAttribute("aria-label", "クイックメニュー");
      dock.innerHTML = dockItems
        .map(([key, href, label]) => `<a class="${key === current ? "active" : ""}" href="${href}"><span>${label}</span></a>`)
        .join("");
      document.body.appendChild(dock);
    }

    document.querySelectorAll(".home-tile, .panel, .boss-unit, .deck-card").forEach((node, index) => {
      node.style.setProperty("--js-order", index % 12);
      node.classList.add("js-polished-surface");
    });

    document.querySelectorAll(".home-tile").forEach((tile, index) => {
      tile.style.setProperty("--tile-hue", String((index * 31) % 360));
    });

    document.querySelectorAll(".home-tile").forEach((tile) => {
      if (tile.querySelector(".js-tile-sheen")) return;
      const sheen = document.createElement("span");
      sheen.className = "js-tile-sheen";
      sheen.setAttribute("aria-hidden", "true");
      tile.appendChild(sheen);
    });

    decorateGameIcons();
  }

  function normalizeText(value) {
    return String(value || "").replace(/\s+/g, "").trim().toUpperCase();
  }

  const generatedIconMap = {
    home: "/static/img/ui/icon-home.png?v=relic-boss-drops-deck-icons-20260630",
    player: "/static/img/ui/icon-player.png?v=relic-boss-drops-deck-icons-20260630",
    deck: "/static/img/ui/icon-deck.png?v=relic-boss-drops-deck-icons-20260630",
    boss: "/static/img/ui/icon-boss.png?v=relic-boss-drops-deck-icons-20260630",
    checkpoint: "/static/img/ui/icon-checkpoint.png?v=relic-boss-drops-deck-icons-20260630",
    gacha: "/static/img/ui/icon-gacha.png?v=relic-boss-drops-deck-icons-20260630",
  };

  function iconForElement(element) {
    const text = normalizeText(element.textContent);
    const href = element.getAttribute("href") || "";
    const route = href.split("?")[0];

    if (element.classList.contains("page-back") || text.includes("戻る")) return "<";
    if (route.endsWith("/index.html") || route === "/" || text.includes("HOME") || text.includes("ホーム")) return "home";
    if (route.endsWith("/mypage.html") || text.includes("PLAYER") || text.includes("マイページ")) return "player";
    if (route.endsWith("/deck.html") || text.includes("DECK") || text.includes("デッキ")) return "deck";
    if (route.endsWith("/boss.html") || text.includes("BOSS") || text.includes("ボス")) return "boss";
    if (route.endsWith("/checkpoint.html") || text.includes("NODE") || text.includes("チェック")) return "checkpoint";
    if (route.endsWith("/gacha.html") || text.includes("GACHA") || text.includes("ガチャ")) return "gacha";
    if (route.endsWith("/cards.html") || route.endsWith("/encyclopedia.html") || text.includes("CARD") || text.includes("カード") || text.includes("図鑑")) return "deck";
    if (text.includes("ログイン")) return "IN";
    if (text.includes("ログアウト") || text.includes("LOGOUT")) return "X";
    if (text.includes("新規") || text.includes("登録")) return "+";
    if (text.includes("かんたん")) return ">>";
    if (text.includes("開始") || text.includes("START") || text.includes("READY")) return ">";
    if (text.includes("再読込") || text.includes("更新") || text.includes("SYNC") || text.includes("REFRESH")) return "R";
    if (text.includes("保存")) return "OK";
    if (text.includes("強化")) return "UP";
    if (text.includes("報酬") || text.includes("DAILY")) return "!";
    if (text.includes("編成")) return "D";
    if (text.includes("探索")) return "S";
    if (text.includes("もう一戦") || text.includes("再挑戦")) return ">";
    return "";
  }

  function addButtonIcon(element, icon) {
    if (!icon || element.querySelector(":scope > .game-btn-icon")) return;
    const iconNode = document.createElement("span");
    iconNode.className = "game-btn-icon";
    iconNode.setAttribute("aria-hidden", "true");
    if (generatedIconMap[icon]) {
      iconNode.style.setProperty("--game-icon-image", `url("${generatedIconMap[icon]}")`);
      iconNode.classList.add("generated-game-icon");
    } else {
      iconNode.textContent = icon;
    }
    element.prepend(iconNode);
    element.classList.add("has-game-icon");
  }

  function addTileIcon(tile, icon) {
    if (!icon || tile.querySelector(":scope > .tile-icon")) return;
    const iconNode = document.createElement("span");
    iconNode.className = "tile-icon";
    iconNode.setAttribute("aria-hidden", "true");
    if (generatedIconMap[icon]) {
      iconNode.style.setProperty("--game-icon-image", `url("${generatedIconMap[icon]}")`);
      iconNode.classList.add("generated-game-icon");
    } else {
      iconNode.textContent = icon;
    }
    tile.prepend(iconNode);
    tile.classList.add("has-game-icon");
  }

  function decorateGameIcons(root = document) {
    root.querySelectorAll("button, .ghost-link, .button-link, .page-tab, .js-bottom-dock a").forEach((element) => {
      addButtonIcon(element, iconForElement(element));
    });
    root.querySelectorAll(".home-tile").forEach((tile) => {
      addTileIcon(tile, iconForElement(tile));
    });
  }

  window.addEventListener("DOMContentLoaded", () => {
    buildDynamicStage();
    document.body.appendChild(overlay);
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        mutation.addedNodes.forEach((node) => {
          if (node.nodeType !== 1) return;
          if (node.matches?.("button, .ghost-link, .button-link, .page-tab, .js-bottom-dock a, .home-tile")) {
            if (node.classList.contains("home-tile")) addTileIcon(node, iconForElement(node));
            else addButtonIcon(node, iconForElement(node));
          }
          decorateGameIcons(node);
        });
      });
    });
    observer.observe(document.body, { childList: true, subtree: true });
    window.decorateGameIcons = decorateGameIcons;
    requestAnimationFrame(() => document.documentElement.classList.remove("is-app-loading"));
  });

  window.gameTransition = {
    enter() {
      document.documentElement.classList.remove("is-app-leaving");
      document.documentElement.classList.add("is-app-entering");
      window.setTimeout(() => document.documentElement.classList.remove("is-app-entering"), 420);
    },
    leave(callback) {
      document.documentElement.classList.add("is-app-leaving");
      window.setTimeout(callback, 260);
    },
  };

  document.addEventListener("click", (event) => {
    const link = event.target.closest("a[href]");
    if (!link || link.target || link.hasAttribute("download")) return;

    const url = new URL(link.href, window.location.href);
    if (url.origin !== window.location.origin) return;
    if (url.href === window.location.href) return;

    event.preventDefault();
    window.gameTransition.leave(() => {
      window.location.href = url.href;
    });
  });

  window.addEventListener("pageshow", () => {
    window.gameTransition.enter();
  });
})();
