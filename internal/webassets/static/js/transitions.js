const state = {
  token: localStorage.getItem("retro_token") || "",
  cardArchiveFilter: "all",
  cardArchiveEntries: [],
  initializing: [],
  onCloseDialog: [],
};

// 共通機能
function appBasePath() {
  const meta = document.querySelector('meta[name="app-base-path"]');
  const raw = window.__APP_BASE_PATH__ || meta?.content || "";
  if (!raw || raw === "/") return "";
  return String(raw).replace(/\/$/, "");
}
function appUrl(path) {
  if (!path || !String(path).startsWith("/")) return path;
  const base = appBasePath();
  if (!base || String(path).startsWith(base + "/")) return path;
  return `${base}${path}`;
}
function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}
const qs = (selector) => document.querySelector(selector);
const qsa = (selector) => [...document.querySelectorAll(selector)];
function setText(selector, value) {
  const node = qs(selector);
  if (node) node.textContent = value;
}
function setHTML(selector, value) {
  const node = qs(selector);
  if (node) node.innerHTML = value;
}
// 新UIのための上書き対応
function setTransition(player, isBattle) {
  const overlay = document.createElement("div");
  overlay.className = "screen-transition";
  overlay.innerHTML = '<div class="transition-core"></div>';
  document.documentElement.classList.add("is-app-loading");

  function buildDynamicStage() {
    document.title = "";
    document.documentElement.classList.add("js-landscape-game-ui");

    if (!isBattle && !document.querySelector(".js-game-hud")) {
      const hud = document.createElement("div");
      hud.className = "js-game-hud";
      hud.innerHTML = `
        <div class="js-game-brand">RELIC RAID</div>
        <div class="js-game-status">
          <span class="js-resource">LV：${player.level}(${player.exp})</span>
          <span class="js-resource">★：<span id="updatableCoins">${player.coins}</span></span>
          <span class="js-resource">${player.username}</span>
        </div>
      `;
      document.body.appendChild(hud);
    }

    if (!isBattle && !document.querySelector(".js-bottom-dock")) {
      const current = document.body.dataset.page || "home";
      const dockItems = [
        ["home", "/static/index.html", "HOME"],
        ["deck", "/static/cards.html", "DECK"],
        ["boss", "/static/boss.html", "BOSS"],
        ["checkpoint", "/static/checkpoint.html", "NODE"],
        ["gacha", "/static/gacha.html", "GACHA"],
        ["encyclopedia", "/static/encyclopedia.html", "LIST"],
      ];
      const dock = document.createElement("nav");
      dock.className = "js-bottom-dock";
      dock.innerHTML = dockItems
        .map(
          ([key, href, label]) =>
            `<a class="${key === current ? "active" : ""}" href="${appUrl(href)}"><span>${label}</span></a>`,
        )
        .join("");
      document.body.appendChild(dock);
    }

    // ダイアログ
    if (true) {
      const hud = document.createElement("div");
      hud.id = "dialogOverlay";
      hud.className = "card-preview-overlay hidden";
      hud.innerHTML = `
        <div id="dialogCloser" style="position: absolute; top: 0px; left: 0px; right: 0px; bottom: 0px; cursor: pointer"></div>
        <button id="dialogCloseButton" class="card-preview-close" type="button">×</button>
        <div id="cardPreviewPanel" class="card-preview-dialog"></div>
      `;
      document.body.appendChild(hud);
      qs("#dialogCloser")?.addEventListener("click", closeDialog);
      qs("#dialogCloseButton")?.addEventListener("click", closeDialog);
    }

    document
      .querySelectorAll(".panel, .boss-unit, .deck-card")
      .forEach((node, index) => {
        node.style.setProperty("--js-order", index % 12);
        node.classList.add("js-polished-surface");
      });

    decorateGameIcons();
  }

  function normalizeText(value) {
    return String(value || "").replace(/\s+/g, "").trim().toUpperCase();
  }

  const generatedIconMap = {
    home: appUrl("/static/img/ui/icon-home.png?v=relic-button-icons-fit-20260702"),
    player: appUrl("/static/img/ui/icon-player.png?v=relic-button-icons-fit-20260702"),
    boss: appUrl("/static/img/ui/icon-boss.png?v=relic-button-icons-fit-20260702"),
    checkpoint: appUrl("/static/img/ui/icon-checkpoint.png?v=relic-button-icons-fit-20260702"),
    gacha: appUrl("/static/img/ui/icon-gacha.png?v=relic-button-icons-fit-20260702"),
    encyclopedia: appUrl("/static/img/ui/icon-encyclopedia.svg?v=relic-button-icons-fit-20260702"),
    deck: appUrl("/static/img/ui/icon-deck.png?v=relic-button-icons-fit-20260702"),
    cardManage: appUrl("/static/img/ui/icon-deck.png?v=relic-button-icons-fit-20260702"),
  };

  function iconForElement(element) {
    const text = normalizeText(element.textContent);
    const href = element.getAttribute("href") || "";
    const route = href.split("?")[0];

    if (
      route.endsWith("/index.html") ||
      route === "/" ||
      text.includes("HOME") ||
      text.includes("ホーム")
    )
      return "home";
    // return "player";
    if (
      route.endsWith("/cards.html") ||
      text.includes("DECK") ||
      text.includes("デッキ")
    )
      return "deck";
    if (
      route.endsWith("/boss.html") ||
      text.includes("BOSS") ||
      text.includes("ボス")
    )
      return "boss";
    if (
      route.endsWith("/checkpoint.html") ||
      text.includes("NODE") ||
      text.includes("チェック")
    )
      return "checkpoint";
    if (
      route.endsWith("/gacha.html") ||
      text.includes("GACHA") ||
      text.includes("ガチャ")
    )
      return "gacha";
    if (
      route.endsWith("/encyclopedia.html") ||
      text.includes("図鑑") ||
      text.includes("詳細確認")
    )
      return "encyclopedia";
    if (
      route.endsWith("/cards.html") ||
      text.includes("カード管理") ||
      text.includes("CARD")
    )
      return "cardManage";
    if (text.includes("カード"))
      return "deck";
    if (text.includes("ログイン")) return "IN";
    if (text.includes("ログアウト") || text.includes("LOGOUT")) return "X";
    if (text.includes("新規") || text.includes("登録")) return "+";
    if (text.includes("かんたん")) return ">>";
    if (
      text.includes("再読込") ||
      text.includes("更新") ||
      text.includes("SYNC") ||
      text.includes("REFRESH")
    )
      return "R";
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
    if (generatedIconMap[icon]) {
      iconNode.style.setProperty(
        "--game-icon-image",
        `url("${generatedIconMap[icon]}")`,
      );
      iconNode.classList.add("generated-game-icon");
    } else {
      iconNode.textContent = icon;
    }
    element.prepend(iconNode);
    element.classList.add("has-game-icon");
  }

  function decorateGameIcons(root = document) {
    root
      .querySelectorAll(
        "button, .js-bottom-dock a",
      )
      .forEach((element) => {
        addButtonIcon(element, iconForElement(element));
      });
  }

  function constentLoaded() {
    buildDynamicStage();
    document.body.appendChild(overlay);
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        mutation.addedNodes.forEach((node) => {
          if (node.nodeType !== 1) return;
          if (
            node.matches?.(
              "button, .js-bottom-dock a",
            )
          ) {
            addButtonIcon(node, iconForElement(node));
          }
          decorateGameIcons(node);
        });
      });
    });
    observer.observe(document.body, { childList: true, subtree: true });
    window.decorateGameIcons = decorateGameIcons;
    requestAnimationFrame(() =>
      document.documentElement.classList.remove("is-app-loading"),
    );
  }

  window.gameTransition = {
    enter() {
      document.documentElement.classList.remove("is-app-leaving");
      document.documentElement.classList.add("is-app-entering");
      window.setTimeout(
        () => document.documentElement.classList.remove("is-app-entering"),
        420,
      );
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

  constentLoaded();
}

function setAuth(token) {
  localStorage.setItem("retro_token", token);
}

function clearAuth() {
  localStorage.removeItem("retro_token");
}

function showToast(message, type = "") {
  const toast = qs("#toast");
  if (!toast) return;
  toast.textContent = message;
  toast.className = `toast ${type}`.trim();
  toast.classList.remove("hidden");
  clearTimeout(showToast.timer);
  showToast.timer = setTimeout(() => toast.classList.add("hidden"), 2800);
}

function rarityStars(rarity) {
  return `<span class="rarity">${"★".repeat(Math.max(1, rarity || 1))}</span>`;
}

function elementLabel(element) {
  return (
    {
      heart: "心",
      tech: "技",
      body: "体",
      none: "無",
    }[element] || element || "無"
  );
}

function cardImageUrl(cardID) {
  return appUrl(`/static/img/cards/card-${String(cardID).padStart(2, "0")}.png?v=relic-button-icons-fit-20260702`);
}

function bossImageUrl(bossID) {
  const id = Number(bossID) || 1;
  const file = `boss-${String(id).padStart(2, "0")}.png`;
  return appUrl(`/static/img/bosses/${file}?v=relic-button-icons-fit-20260702`);
}

function subElementLabel(effect) {
  return (
    {
      tentacle: "触腕",
      abyss: "深淵",
      fire: "火",
      storm: "風",
      holy: "光",
      void: "闇",
      quake: "地震",
      venom: "毒",
      ice: "氷",
      fang: "牙",
      spike: "棘",
      cosmic: "星辰",
      water: "水",
    }[effect] ||
    effect ||
    "無"
  );
}

function skillGraphicHTML(kind = "none") {
  const safeKind = String(kind || "none").replace(/[^a-z0-9_-]/gi, "");
  return `<span class="skill-graphic skill-graphic-${safeKind}"><i></i></span>`;
}

function setLoading(show) {
  if (show) {
    document.documentElement.classList.add("is-app-loading");
  } else {
    document.documentElement.classList.remove("is-app-loading");
  }
}

async function api(path, options = {}) {
  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {}),
  };
  if (state.token) headers.Authorization = `Bearer ${state.token}`;

  const res = await fetch(appUrl(path), { ...options, headers });
  const contentType = res.headers.get("content-type") || "";
  const data = contentType.includes("application/json")
    ? await res.json()
    : await res.text();

  if (!res.ok) {
    const message = data?.error || `request failed: ${res.status}`;
    if (res.status === 401) {
      clearAuth();
      location.href = appUrl("/static/index.html");
    }
    throw new Error(message);
  }

  return data;
}

function activateTab(name) {
  qsa(".tab").forEach((tab) =>
    tab.classList.toggle("active", tab.dataset.tab === name),
  );
  qsa("[data-tab-body]").forEach((body) =>
    body.classList.toggle("hidden", body.dataset.tabBody !== name),
  );
}

function escapeHTML(value) {
  return String(value ?? "").replace(
    /[&<>"']/g,
    (ch) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
      ch
      ],
  );
}

function cardSkillHTML(skill) {
  if (!skill?.name) return "";
  const targets = { "target": "攻撃対象", "all": "全体" };
  return `
    <article class="card-preview-skill">
      <div class="checkpoint-tag-row">
        <span>発動率：${skill.trigger_rate}%</span>
        ${skill.sub_element ? `<span>副属性：${subElementLabel(skill.sub_element)}のみ</span>` : ""}
        <span>対象：${targets[skill.target] ?? "不明"}</span>
        ${skill.heal ? `<span>回復量${skill.heal}</span>` : ""}
        ${skill.shield ? `<span>最大${skill.shield}の盾</span>` : ""}
        ${skill.evade ? `<span>1回無敵</span>` : ""}
        ${skill.first ? `<span>先制</span>` : ""}
        ${skill.damaged ? `<span>被ダメージ時のみ</span>` : ""}
        ${skill.dead ? `<span>無効化時のみ</span>` : ""}
      </div>
      <div>${skillGraphicHTML(skill.effect_type || "none")}<strong>${escapeHTML(skill.name)}</strong></div>
      <p>${escapeHTML(skill.description || "")}</p>
    </article>
  `;
}

// ダイアログを開く
function openDialog(color) {
  const overlay = qs("#dialogOverlay");
  const panel = qs("#cardPreviewPanel");
  if (!overlay || !panel) return null;
  overlay.className = `card-preview-overlay ` + (color || "");
  return panel;
}

// ダイアログを閉じる
async function closeDialog() {
  qs("#dialogOverlay")?.classList.add("hidden");
  for (let func of state.onCloseDialog) await func();
}

// ミニカードのHTMLを生成する
function renderCardItem(a, obtained, status, pointer, buttons) {
  const card = a || {};
  const statusText = status ? `<span class="card-archive-status">${status}</span>` : "";
  let buttonsText = "";
  if (buttons && buttons.length) {
    buttonsText += `<div class="actions-row">`;
    for (let i = 0; i < buttons.length; i++) {
      buttonsText += `<button id="cardItemButton${i}" data-card-id="${card.id}">${buttons[i]}</button>`;
    }
    buttonsText += `</div>`;
  }
  return `<article class="card-archive-entry ${obtained ? "owned" : "missing"} element-${card.element || "none"}" ${pointer ? "style='cursor: pointer;'" : ""} data-card-id="${card.id}" role="button" tabindex="0">
    <div class="card-archive-art">
      <img src="${cardImageUrl(card.id)}" loading="lazy" />
      <div class="card-archive-rare">${rarityStars(card.rarity)}</div>
      ${statusText}
      <div class="card-archive-element card-preview-kicker">${elementLabel(card.element)}</div>
      <div class="card-archive-meta">${card.max_hp} / ${card.defense}</div>
    </div>
    <div class="card-archive-body">
      <div class="card-archive-title"><strong>${card.name}（${card.sub_name}）</strong></div>
    </div>
    ${buttonsText}
  </article>`;
}

// カードダイアログのHTMLを生成する
function openCardDialog(card, status) {
  if (!card) return;

  const panel = openDialog(`element-${card.element || "none"}`);
  if (!panel) return;

  const statusText = status ? `<span class="card-archive-status">${status}</span>` : "";
  const skills = card.skills.map(cardSkillHTML).join("");
  panel.innerHTML=`
      <div class="boss-detail-hero">
        <div class="card-preview-art">
          <img src="${cardImageUrl(card.id)}" loading="lazy" />
          ${statusText}
        </div>
        <div>
          <div style="display: flex; gap: 8px"><div class="card-preview-kicker">${elementLabel(card.element)}</div>${rarityStars(card.rarity)}</div>
          <h2>${card.name}（${card.sub_name}）</h2>
          <p>${card.description || ""}</p>
        </div>
      </div>
      <div class="card-preview-info">
        <div class="card-preview-stats">
          <span>HP ${card.max_hp ?? "-"}</span>
          <span>DEF ${card.defense ?? "-"}</span>
        </div>
        <div class="card-preview-skills">${skills}</div>
      </div>
  `;
}

async function loadPlayer() {
    if (!state.token) return null;
    const player = await api("/api/player/me");
    setText("#updatableCoins", player.coins);
    return player;
}

// 初期化
async function init() {
  qsa(".tab").forEach((tab) =>
    tab.addEventListener("click", () => activateTab(tab.dataset.tab)),
  );

  setLoading(true);
  try {
    const player = await loadPlayer();
    if (!player && !qs("#authView")) location.href = appUrl("/static/index.html");

    // チュートリアル
    let message = null;
    if (player && player.tutorial) {
      if (player.owned == 2) {
        if (document.body.dataset.page == "home") {
          message = "このゲームは防災キャラクターカードを集めて編成し、災害ボスの攻撃を耐えきることが目的です\r\n早速ボスに挑んでみましょう\r\n右側の「BOSS」を選択してください";
        } else if (document.body.dataset.page == "boss") {
          message = "現在挑戦できるボスは「防災訓練ゴーレム１」だけです\r\nこのボスは、順番に3回攻撃してきます\r\n1. 心属性威力50「災害の噂」\r\n2. 技属性威力50「避難警報」\r\n3. 体属性威力50「床上浸水」\r\nボスのアイコンをクリックして、ダイアログを開き「このボスに挑戦」を選択してください";
        } else if (document.body.dataset.page == "battle") {
          message = "戦闘が始まりました。\r\nデッキの「蓄電核ボルト」のHPは100です。\r\nボスは威力50で3回攻撃してくるので、このままでは耐久出来ません\r\nただし「蓄電核ボルト」は回復スキルを持っているので、攻撃を受けた時に1回だけHPを30回復できます\r\nまた「蓄電核ボルト」は体属性で防御力を30持っているので、ボスの3回目の攻撃を20に減らすことができます\r\nこれで耐久出来るはずなので「開始」ボタンを選択してください";
        } else {
          location.href = appUrl("/static/index.html");
        }
      }
      if (player.owned == 3 && player.decks <= 2) {
        if (document.body.dataset.page == "boss") {
          message = "ボスに勝利すると、カードを入手できる場合があります\r\n今回は「夢避けチャーム」を入手しました\r\n右側の「DECK」を選択して、さっそく編成しましょう";
        } else if (document.body.dataset.page == "deck") {
          message = "もともと所持していた「地図鱗マップ」と、入手した「夢避けチャーム」をデッキに編成します\r\nそれぞれの「追加」ボタンをクリックし、全てのカードをデッキに追加してください\r\n全て編成できたら、右側の「BOSS」を選択して次のボスに挑みましょう";
        } else {
          location.href = appUrl("/static/boss.html");
        }
      }
      if (player.owned == 3 && player.decks == 3) {
        if (document.body.dataset.page == "boss") {
          message = "次に挑戦するボスは「防災訓練ゴーレム２」です\r\nボスのアイコンをクリックすると、ボスの攻撃や入手できるカードが確認できます\r\nボスの3回目の攻撃の副属性が水であることを確認してから、「このボスに挑戦」を選択してください";
        } else if (document.body.dataset.page == "battle") {
          message = "各カードは戦闘中に一回だけ、スキルを発動することができます\r\n「蓄電核ボルト」のスキルは、攻撃を受けた後にHPを30回復します\r\n「夢避けチャーム」のスキルは、攻撃を受ける前にダメージを30軽減する盾を作ります\r\n「地図鱗マップ」のスキルは攻撃を1回無効化する強力なものですが、攻撃の副属性が水でないと発動しません\r\nこのボスは3回目に水属性の攻撃をするので、無効化できるはずです";
        } else {
          location.href = appUrl("/static/boss.html");
        }
      }
      if (player.owned == 4 && player.decks == 3) {
          message = "これでチュートリアルは終了です\r\n入手した防災キャラクターカードを用いて、他のボスにも挑戦してみましょう";
      }
    }

    // 全ての画面を初期化
    const isBattle = !state.token || document.body.classList.contains("battle-only");
    setTransition(player, isBattle);
    for (let func of state.initializing) await func(player);

    // チュートリアルメッセージ表示
    if (message) {
      let panel = openDialog();
      panel.innerHTML = `<div style="white-space: pre-wrap;">${message}</div>`
    }
  } finally {
    setLoading(false);
  }
}

window.addEventListener("DOMContentLoaded", () => {
  init().catch((err) => showToast(err.message, "error"));
});