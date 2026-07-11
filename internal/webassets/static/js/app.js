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
function staticUrl(path) {
  return appUrl(path);
}
const state = {
  token: localStorage.getItem("retro_token") || "",
  userId: Number(localStorage.getItem("retro_user_id") || "0"),
  username: localStorage.getItem("retro_username") || "",
  cardArchiveFilter: "all",
  cardArchiveEntries: [],
};

const ANDROID_AUTO_USER_KEY = "retro_android_auto_username";
const ANDROID_AUTO_PASS_KEY = "retro_android_auto_password";

const qs = (selector) => document.querySelector(selector);
const qsa = (selector) => [...document.querySelectorAll(selector)];

function setText(selector, value) {
  const node = qs(selector);
  if (node) node.textContent = value;
}

function syncShellPlayerName(username = state.username) {
  const shellPlayer = qs(".js-player");
  if (shellPlayer) shellPlayer.textContent = username || "PLAYER";
}

function setAuth(token, userId, username = "") {
  state.token = token;
  state.userId = userId;
  state.username = username;
  localStorage.setItem("retro_token", token);
  localStorage.setItem("retro_user_id", String(userId));
  if (username) localStorage.setItem("retro_username", username);
  syncShellPlayerName(username);
}

function clearAuth() {
  state.token = "";
  state.userId = 0;
  state.username = "";
  localStorage.removeItem("retro_token");
  localStorage.removeItem("retro_user_id");
  localStorage.removeItem("retro_username");
  syncShellPlayerName("GUEST");
}

function showToast(message, type = "") {
  const toast = qs("#toast");
  toast.textContent = message;
  toast.className = `toast ${type}`.trim();
  toast.classList.remove("hidden");
  clearTimeout(showToast.timer);
  showToast.timer = setTimeout(() => toast.classList.add("hidden"), 2800);
}

function rarityStars(rarity) {
  return `<span class="rarity">${"★".repeat(Math.max(1, rarity || 1))}</span>`;
}

function portraitClass(frame) {
  if (frame === "gold") return "card-frame-gold";
  if (frame === "boss") return "card-frame-boss";
  return "card-frame-bronze";
}

function elementLabel(element) {
  return (
    {
      fire: "火",
      water: "水",
      earth: "木",
      wind: "風",
      light: "光",
      dark: "闇",
      heart: "心",
      tech: "技",
      body: "体",
      wood: "木",
      none: "無",
    }[element || "none"] ||
    element ||
    "無"
  );
}

function cardImageUrl(cardID) {
  return staticUrl(`/static/img/cards/card-${String(cardID).padStart(2, "0")}.png?v=relic-button-icons-fit-20260702`);
}

function bossImageUrl(bossID = 1) {
  const id = Number(bossID) || 1;
  const files = {
    1: "golem.png",
    2: "leviathan.png",
    3: "ifrit.png",
    4: "garuda.png",
    5: "seraphim.png",
    6: "nightmare.png",
  };
  const file = files[id] || `boss-${String(id).padStart(2, "0")}.png`;
  return staticUrl(`/static/img/bosses/${file}?v=relic-button-icons-fit-20260702`);
}

function bossEffectLabel(effect) {
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
    "特殊"
  );
}

function skillGraphicHTML(kind = "none") {
  const safeKind = String(kind || "none").replace(/[^a-z0-9_-]/gi, "");
  return `<span class="skill-graphic skill-graphic-${safeKind}" aria-hidden="true"><i></i></span>`;
}

function bossHintList(items = [], renderer = (item) => escapeHTML(item)) {
  return items.length
    ? items.map((item) => `<li>${renderer(item)}</li>`).join("")
    : "<li>情報解析中</li>";
}

function setLoading(show) {
  qs("#loadingOverlay")?.classList.toggle("hidden", !show);
}

function isAndroidAppWebView() {
  const params = new URLSearchParams(location.search);
  return (
    /RelicRaidAndroid/i.test(navigator.userAgent) ||
    params.get("android_app") === "1"
  );
}

function createAndroidCredentialPair() {
  const random = Math.random().toString(36).slice(2, 10);
  const time = Date.now().toString(36).slice(-6);
  return {
    username: `android_${time}_${random}`,
    password: `androidpass_${time}_${random}`,
  };
}

function getAndroidCredentials() {
  let username = localStorage.getItem(ANDROID_AUTO_USER_KEY) || "";
  let password = localStorage.getItem(ANDROID_AUTO_PASS_KEY) || "";
  if (!username || !password) {
    const next = createAndroidCredentialPair();
    username = next.username;
    password = next.password;
    localStorage.setItem(ANDROID_AUTO_USER_KEY, username);
    localStorage.setItem(ANDROID_AUTO_PASS_KEY, password);
  }
  return { username, password };
}

function replaceAndroidCredentials() {
  const next = createAndroidCredentialPair();
  localStorage.setItem(ANDROID_AUTO_USER_KEY, next.username);
  localStorage.setItem(ANDROID_AUTO_PASS_KEY, next.password);
  return next;
}

function updateClock() {
  const clock = qs("#clockText");
  if (!clock) return;
  clock.textContent = new Date().toLocaleTimeString("ja-JP", {
    hour: "2-digit",
    minute: "2-digit",
  });
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
      renderAuthState();
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

function bindTabs() {
  qsa(".tab").forEach((tab) =>
    tab.addEventListener("click", () => activateTab(tab.dataset.tab)),
  );
}

function currentPage() {
  return document.body.dataset.page || "home";
}

function syncNav() {
  const page = currentPage();
  qsa(".page-tab").forEach((tab) => {
    const active = tab.dataset.route === page;
    tab.classList.toggle("active", active);
    tab.setAttribute("aria-current", active ? "page" : "false");
  });
}

async function login(username, password) {
  setLoading(true);
  try {
    const data = await api("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
    setAuth(data.token, data.user_id, username);
    showToast("ログインしました", "success");
    await bootstrapGame();
  } finally {
    setLoading(false);
  }
}

async function registerAndLogin(username, password) {
  setLoading(true);
  try {
    const data = await api("/api/auth/register", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
    setAuth(data.token, data.user_id, username);
    showToast("登録完了、そのまま開始します", "success");
    await bootstrapGame();
  } finally {
    setLoading(false);
  }
}

async function quickStart() {
  const suffix = Date.now().toString().slice(-6);
  await registerAndLogin(`player_${suffix}`, `startpass${suffix}`);
}

async function authenticateAndroidCredentials(credentials) {
  try {
    const data = await api("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(credentials),
    });
    return data;
  } catch (loginErr) {
    try {
      const data = await api("/api/auth/register", {
        method: "POST",
        body: JSON.stringify(credentials),
      });
      return data;
    } catch (registerErr) {
      const fallback = replaceAndroidCredentials();
      const data = await api("/api/auth/register", {
        method: "POST",
        body: JSON.stringify(fallback),
      });
      credentials.username = fallback.username;
      credentials.password = fallback.password;
      return data;
    }
  }
}

async function autoLoginForAndroid() {
  if (!isAndroidAppWebView() || state.token) return false;
  if (!qs("#authView")) {
    location.href = appUrl("/static/index.html");
    return true;
  }

  setLoading(true);
  try {
    const credentials = getAndroidCredentials();
    const data = await authenticateAndroidCredentials(credentials);
    setAuth(data.token, data.user_id, credentials.username);
    showToast("Android自動ログインしました", "success");
    await bootstrapGame();
    return true;
  } catch (err) {
    renderAuthState();
    showToast(`自動ログインに失敗しました: ${err.message}`, "error");
    return false;
  } finally {
    setLoading(false);
  }
}

function bindAuthForms() {
  qs("#loginForm")?.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      await login(form.get("username"), form.get("password"));
    } catch (err) {
      showToast(err.message, "error");
    }
  });

  qs("#registerForm")?.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      await registerAndLogin(form.get("username"), form.get("password"));
    } catch (err) {
      showToast(err.message, "error");
    }
  });

  qs("#quickStartBtn")?.addEventListener("click", async () => {
    try {
      await quickStart();
    } catch (err) {
      showToast(err.message, "error");
    }
  });
}

function renderAuthState() {
  const ok = Boolean(state.token);
  const allowMypagePreview =
    currentPage() === "mypage" && Boolean(qs(".mypage-hud"));
  const authView = qs("#authView");
  const gameView = qs("#gameView");
  authView?.classList.toggle("hidden", ok);
  gameView?.classList.toggle("hidden", !ok && !allowMypagePreview);
  syncShellPlayerName(ok ? state.username : "GUEST");
  const badge = qs("#authBadge");
  if (!badge) return;
  badge.classList.remove("hidden");
  badge.textContent = ok
    ? `ログイン中: ${state.username || "player"}`
    : "未ログイン";
}

function renderPlayer(player) {
  const fields = {
    playerName: player.username || state.username,
    playerLevel: player.level,
    playerExp: player.exp,
    playerStamina: "∞",
    playerCoins: player.coins,
    playerExplorations: player.total_explorations ?? 0,
  };
  Object.entries(fields).forEach(([id, value]) => {
    const node = qs(`#${id}`);
    if (node) node.textContent = value;
  });
}

function renderAreas(areas) {
  const root = qs("#areasList");
  if (!root) return;
  root.innerHTML = areas
    .map(
      (area) => `
        <article class="card">
          <h3>${area.name}</h3>
          <div class="meta">必要Lv: ${area.required_level} / 消費スタミナ: ∞</div>
          <p class="meta">${area.description || ""}</p>
          <button data-area-id="${area.id}" data-area-name="${area.name}">探索する</button>
        </article>
      `,
    )
    .join("");

  root.querySelectorAll("button[data-area-id]").forEach((button) =>
    button.addEventListener("click", async () => {
      try {
        await runExploreSequence(button.dataset.areaName);
        setLoading(true);
        const result = await api("/api/explore", {
          method: "POST",
          body: JSON.stringify({ area_id: Number(button.dataset.areaId) }),
        });
        renderExploreResult(result);
        renderPlayer(await api("/api/player/me"));
        await Promise.all([loadInventory(), loadEncyclopedia()]);
      } catch (err) {
        showToast(err.message, "error");
      } finally {
        setLoading(false);
      }
    }),
  );
}

async function runExploreSequence(name) {
  const box = qs("#exploreSequence");
  if (!box) return;
  box.innerHTML = "<div>移動先選択...</div>";
  await new Promise((resolve) => setTimeout(resolve, 220));
  box.innerHTML = `<div>${name} に移動中...</div>`;
  await new Promise((resolve) => setTimeout(resolve, 380));
  box.innerHTML = "<div>探索中...</div>";
  await new Promise((resolve) => setTimeout(resolve, 480));
  box.innerHTML = "<div>結果を取得中...</div>";
}

function renderExploreResult(result) {
  const root = qs("#exploreResult");
  if (!root) return;
  const item = result.new_item
    ? `<p><strong>発見:</strong> ${result.new_item.name} ${rarityStars(result.new_item.rarity)}</p>`
    : "";
  root.innerHTML = `
    <h3>探索結果</h3>
    <p>${result.message}</p>
    ${item}
    <p>EXP +${result.gained_exp} / コイン +${result.gained_coins}</p>
    ${result.encyclopedia_registered ? "<p>図鑑に新規登録されました。</p>" : ""}
  `;
}

async function loadPlayer() {
  renderPlayer(await api("/api/player/me"));
}

async function loadAreas() {
  renderAreas(await api("/api/areas"));
}

async function loadInventory() {
  const root = qs("#inventoryList");
  if (!root) return;
  const inventory = await api("/api/items/inventory");
  root.innerHTML = inventory.length
    ? inventory
        .map(
          (item) =>
            `<div class="list-item"><strong>${item.name}</strong> ${rarityStars(item.rarity)}<div>所持数: ${item.quantity}</div></div>`,
        )
        .join("")
    : '<div class="muted">まだアイテムを持っていません。</div>';
}

async function loadEncyclopedia() {
  const rate = qs("#encyclopediaRate");
  const list = qs("#encyclopediaList");
  if (!rate || !list) return;
  const data = await api("/api/cards/archive");
  state.cardArchiveEntries = data.cards || [];
  rate.textContent = `収集率: ${data.obtained_count ?? 0} / ${data.total ?? 0} (${(data.completion_rate ?? 0).toFixed(1)}%)`;
  renderCardArchive();
}

function renderCardArchive() {
  const list = qs("#encyclopediaList");
  if (!list) return;
  const filter = state.cardArchiveFilter || "all";
  const entries = (state.cardArchiveEntries || []).filter((entry) => {
    if (filter === "owned") return entry.obtained;
    if (filter === "missing") return !entry.obtained;
    if (filter === "all") return true;
    return entry.card?.element === filter;
  });
  list.classList.toggle("muted", !entries.length);
  list.innerHTML = entries.length
    ? entries
        .map((entry) => {
          const card = entry.card || {};
          const status = entry.obtained
            ? entry.in_deck
              ? `デッキ S${entry.deck_slot}`
              : "所持済み"
            : "未所持";
          return `
            <article class="card-archive-entry ${entry.obtained ? "owned" : "missing"} element-${card.element || "none"}" data-card-id="${card.id}" role="button" tabindex="0" aria-label="${card.name}を拡大表示">
              <div class="card-archive-art">
                <img src="${cardImageUrl(card.id)}" alt="${card.name}" loading="lazy" />
                <span class="card-archive-status">${status}</span>
              </div>
              <div class="card-archive-body">
                <div class="card-archive-title"><strong>${card.name}</strong>${rarityStars(card.rarity)}</div>
                <div class="card-archive-meta"><span class="element-badge element-${card.element || "none"}">${elementLabel(card.element)}</span><span>HP ${card.max_hp}</span><span>ATK ${card.attack}</span><span>DEF ${card.defense}</span></div>
                <p>${card.description || ""}</p>
              </div>
            </article>
          `;
        })
        .join("")
    : '<div class="list-item">該当するカードがありません。</div>';

  list
    .querySelectorAll(".card-archive-entry[data-card-id]")
    .forEach((entryNode) => {
      entryNode.addEventListener("click", () =>
        openCardPreview(Number(entryNode.dataset.cardId)),
      );
      entryNode.addEventListener("keydown", (event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          openCardPreview(Number(entryNode.dataset.cardId));
        }
      });
    });
}

function archiveEntryByCardID(cardID) {
  return (state.cardArchiveEntries || []).find(
    (entry) => entry.card?.id === cardID,
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

function cardSkillHTML(title, skill, tone = "", graphic = "none") {
  if (!skill?.name) return "";
  const rate = Number(skill.trigger_rate || 0);
  const rateText =
    rate > 0 && rate < 100
      ? `<span class="skill-rate">発動 ${rate}%</span>`
      : "";
  const effectText = skill.effect_type
    ? `<span class="skill-rate">${skillEffectLabel(skill.effect_type)}</span>`
    : "";
  return `
    <article class="card-preview-skill ${tone}">
      <div><span class="skill-label">${title}</span><span>${effectText}${rateText}</span></div>
      <div class="skill-visual-row">${skillGraphicHTML(skill.effect_type || graphic)}<strong>${escapeHTML(skill.name)}</strong></div>
      <p>${escapeHTML(skill.description || "")}</p>
    </article>
  `;
}

function skillEffectLabel(effect) {
  return (
    {
      shield: "盾役",
      heal: "回復補助",
      mitigate: "軽減",
      evade: "回避",
      revive: "蘇生",
      advantage: "特攻",
    }[effect] || "防御"
  );
}

function openCardPreview(cardID) {
  const entry = archiveEntryByCardID(cardID);
  const card = entry?.card || {};
  const overlay = qs("#cardPreviewOverlay");
  if (!overlay || !card.id) return;
  if (overlay.parentElement !== document.body)
    document.body.appendChild(overlay);

  const status = entry.obtained
    ? entry.in_deck
      ? `デッキ S${entry.deck_slot}`
      : "所持済み"
    : "未所持";
  const image = qs("#cardPreviewImage");
  if (image) {
    image.src = cardImageUrl(card.id);
    image.alt = card.name || "カード";
  }
  setText("#cardPreviewStatus", status);
  setText("#cardPreviewElement", elementLabel(card.element));
  setText("#cardPreviewName", card.name || "カード");
  setText("#cardPreviewRarity", "★".repeat(Math.max(1, card.rarity || 1)));
  setText("#cardPreviewHP", `HP ${card.max_hp ?? "-"}`);
  setText("#cardPreviewATK", `ATK ${card.attack ?? "-"}`);
  setText("#cardPreviewDEF", `DEF ${card.defense ?? "-"}`);
  setText("#cardPreviewDescription", card.description || "");
  const skills = qs("#cardPreviewSkills");
  if (skills) {
    skills.innerHTML = [
      cardSkillHTML(
        "固有スキル",
        card.unique_skill,
        "unique-skill",
        card.element || "none",
      ),
      cardSkillHTML(
        "特攻防御",
        card.advantage_defense_skill,
        "advantage-defense",
        "advantage",
      ),
    ].join("");
  }

  overlay.className = `card-preview-overlay element-${card.element || "none"}`;
  overlay.classList.remove("hidden");
  overlay.setAttribute("aria-hidden", "false");
  qs("#closeCardPreviewBtn")?.focus();
}

function closeCardPreview() {
  const overlay = qs("#cardPreviewOverlay");
  if (!overlay) return;
  overlay.classList.add("hidden");
  overlay.setAttribute("aria-hidden", "true");
}

async function loadNotices() {
  const root = qs("#noticesList");
  if (!root) return;
  const notices = await api("/api/notices");
  root.innerHTML = notices
    .map(
      (notice) =>
        `<div class="list-item"><strong>${notice.is_pinned ? "📌 " : ""}${notice.title}</strong><div class="meta">${new Date(notice.published_at).toLocaleString("ja-JP")}</div><div>${notice.body}</div></div>`,
    )
    .join("");
}

async function loadDeckPreview() {
  const root = qs("#deckPreview");
  if (!root) return;
  const data = await api("/api/cards/deck");
  root.innerHTML = data.cards
    .map(
      (deckCard) =>
        `<div class="mini-card">
          <img class="card-thumb" src="${cardImageUrl(deckCard.card.id)}" alt="${deckCard.card.name}" loading="lazy" />
          <div class="mini-card-name">${deckCard.card.name}</div>
          <div class="mini-card-stats">[${deckCard.card.element}] HP ${deckCard.card.max_hp} / ATK ${deckCard.card.attack}</div>
        </div>`,
    )
    .join("");
}

async function loadBossPreview() {
  const root = qs("#bossPreview");
  if (!root) return;
  const data = await api("/api/boss");
  const bosses = data.bosses?.length ? data.bosses : [data.boss];
  state.bosses = bosses;
  state.bossDetails = state.bossDetails || {};
  if (data.boss?.id) {
    state.bossDetails[Number(data.boss.id)] = {
      recommended_deck: data.recommended_deck || [],
      drop_preview: data.drop_preview || null,
    };
  }
  root.innerHTML = bosses
    .map(
      (boss) => `
        <button class="portrait-panel ${portraitClass(boss.frame_style)} boss-list-card" type="button" data-boss-id="${boss.id}">
          <div class="portrait-art boss-art"><img class="boss-preview-image" src="${bossImageUrl(boss.id)}" alt="${boss.name}" loading="lazy" /></div>
          <strong>${boss.name}</strong>
          <div class="meta">属性: ${elementLabel(boss.element)} / HP ${boss.max_hp} / ATK ${boss.attack} / DEF ${boss.defense}</div>
          <div class="meta">勝利報酬: EXP ${boss.reward_exp} / COIN ${boss.reward_coins}</div>
        </button>
      `,
    )
    .join("");
  root.querySelectorAll(".boss-list-card[data-boss-id]").forEach((node) => {
    node.addEventListener("click", () =>
      selectBossDetail(Number(node.dataset.bossId)).catch((err) =>
        showToast(err.message, "error"),
      ),
    );
  });
  renderBossDetail(Number(data.boss?.id || bosses[0]?.id || 1));
}

async function selectBossDetail(bossID) {
  if (!state.bossDetails?.[Number(bossID)]) {
    const data = await api(`/api/boss?id=${bossID}`);
    const boss = data.boss;
    if (boss?.id) {
      const index = (state.bosses || []).findIndex(
        (item) => Number(item.id) === Number(boss.id),
      );
      if (index >= 0) state.bosses[index] = boss;
      state.bossDetails[Number(boss.id)] = {
        recommended_deck: data.recommended_deck || [],
        drop_preview: data.drop_preview || null,
      };
    }
  }
  renderBossDetail(bossID);
}

function bossDropPreviewHTML(dropPreview) {
  if (!dropPreview) return "";
  const candidates = dropPreview.candidates || [];
  return `
    <section class="boss-drop-panel">
      <div class="boss-strategy-title">ボスドロップ <span>${Number(dropPreview.drop_rate_percent || 0)}%</span></div>
      <div class="boss-drop-list">
        ${
          candidates
            .map(
              (card) => `
              <article class="boss-drop-card">
                <img src="${cardImageUrl(card.id)}" alt="${escapeHTML(card.name)}" loading="lazy" />
                <div>
                  <strong>${escapeHTML(card.name)}</strong>
                  <span>${elementLabel(card.element)} / ★${Number(card.rarity || 1)}</span>
                </div>
              </article>
            `,
            )
            .join("") || `<div class="muted">候補解析中</div>`
        }
      </div>
    </section>
  `;
}

function recommendedDeckHTML(cards = []) {
  return `
    <section class="boss-recommend-panel">
      <div class="boss-strategy-title">おすすめデッキ提案</div>
      <div class="boss-recommend-list">
        ${
          cards
            .map(
              (entry, index) => `
              <article class="boss-recommend-card">
                <span class="recommend-rank">${index + 1}</span>
                <img src="${cardImageUrl(entry.card?.id)}" alt="${escapeHTML(entry.card?.name || "カード")}" loading="lazy" />
                <div>
                  <strong>${escapeHTML(entry.card?.name || "カード")}</strong>
                  <span>${elementLabel(entry.card?.element)} / ${escapeHTML(entry.reason || "耐久補助")}</span>
                </div>
              </article>
            `,
            )
            .join("") ||
          `<div class="muted">所持カードから提案できる候補がありません。</div>`
        }
      </div>
    </section>
  `;
}

function renderBossDetail(bossID) {
  const panel = qs("#bossDetailPanel");
  if (!panel) return;
  const boss = (state.bosses || []).find(
    (item) => Number(item.id) === Number(bossID),
  );
  if (!boss) {
    panel.textContent = "ボス情報を読み込めませんでした。";
    return;
  }
  qsa(".boss-list-card").forEach((node) =>
    node.classList.toggle(
      "active",
      Number(node.dataset.bossId) === Number(bossID),
    ),
  );
  const moves = boss.attack_moves || [];
  const hint = boss.strategy_hint || {};
  const detail = state.bossDetails?.[Number(bossID)] || {};
  panel.classList.remove("muted");
  panel.innerHTML = `
    <div class="boss-detail-hero">
      <img src="${bossImageUrl(boss.id)}" alt="${escapeHTML(boss.name)}" loading="lazy" />
      <div>
        <div class="card-preview-kicker">BOSS DETAIL</div>
        <h3>${escapeHTML(boss.name)}</h3>
        <p>${escapeHTML(boss.description || "")}</p>
      </div>
    </div>
    <div class="boss-detail-stats">
      <span class="element-badge element-${boss.element || "none"}">攻撃属性: ${elementLabel(boss.element)}</span>
      <span>HP ${boss.max_hp}</span>
      <span>ATK ${boss.attack}</span>
      <span>DEF ${boss.defense}</span>
    </div>
    <section class="boss-strategy-grid" aria-label="攻略ヒント">
      <article class="boss-strategy-card">
        <div class="boss-strategy-title">有効属性</div>
        <div class="boss-effective-elements">
          ${
            (hint.effective_elements || [])
              .map(
                (element) =>
                  `<span class="element-badge element-${element}">${elementLabel(element)}</span>`,
              )
              .join("") ||
            `<span class="element-badge element-none">解析中</span>`
          }
        </div>
      </article>
      <article class="boss-strategy-card">
        <div class="boss-strategy-title">注意すべき技</div>
        <ul>${bossHintList(hint.dangerous_moves)}</ul>
      </article>
      <article class="boss-strategy-card">
        <div class="boss-strategy-title">おすすめカード傾向</div>
        <ul>${bossHintList(hint.recommended_cards)}</ul>
      </article>
    </section>
    ${bossDropPreviewHTML(detail.drop_preview)}
    ${recommendedDeckHTML(detail.recommended_deck || [])}
    <div class="boss-move-list">
      ${moves
        .map(
          (move) => `
            <article class="boss-move-item effect-${move.effect_type || "none"}">
              <div class="boss-move-head">${skillGraphicHTML(move.effect_type || "none")}<strong>${escapeHTML(move.name)}</strong><span class="boss-move-type">${bossEffectLabel(move.effect_type)}</span></div>
              <p>攻撃属性 ${elementLabel(boss.element)} / 追加威力 +${Number(move.power_bonus || 0)}</p>
            </article>
          `,
        )
        .join("")}
    </div>
    <a class="button-link boss-challenge-link" href="${appUrl(`/static/battle.html?boss_id=${boss.id}`)}">このボスに挑戦</a>
  `;
}

async function claimBonus() {
  setLoading(true);
  try {
    const data = await api("/api/login-bonus/claim", {
      method: "POST",
      body: "{}",
    });
    showToast(`ログインボーナス獲得: ${data.reward_value} コイン`, "success");
    await loadPlayer();
  } finally {
    setLoading(false);
  }
}

function bindGameActions() {
  qs("#refreshBtn")?.addEventListener("click", () =>
    bootstrapGame().catch((err) => showToast(err.message, "error")),
  );
  qs("#claimBonusBtn")?.addEventListener("click", () =>
    claimBonus().catch((err) => showToast(err.message, "error")),
  );
  qs("#reloadInventoryBtn")?.addEventListener("click", () =>
    loadInventory().catch((err) => showToast(err.message, "error")),
  );
  qs("#reloadEncyclopediaBtn")?.addEventListener("click", () =>
    loadEncyclopedia().catch((err) => showToast(err.message, "error")),
  );
  qsa("[data-card-filter]").forEach((button) =>
    button.addEventListener("click", () => {
      state.cardArchiveFilter = button.dataset.cardFilter || "all";
      qsa("[data-card-filter]").forEach((node) =>
        node.classList.toggle("active", node === button),
      );
      renderCardArchive();
    }),
  );
  qs("#closeCardPreviewBtn")?.addEventListener("click", closeCardPreview);
  qs("#cardPreviewOverlay")?.addEventListener("click", (event) => {
    if (event.target === event.currentTarget) closeCardPreview();
  });
  document.addEventListener("keydown", (event) => {
    if (
      event.key === "Escape" &&
      !qs("#cardPreviewOverlay")?.classList.contains("hidden")
    ) {
      closeCardPreview();
    }
  });
  qs("#reloadNoticesBtn")?.addEventListener("click", () =>
    loadNotices().catch((err) => showToast(err.message, "error")),
  );
  qs("#logoutBtn")?.addEventListener("click", () => {
    clearAuth();
    renderAuthState();
    if (
      location.pathname !== "/" &&
      !location.pathname.endsWith("/index.html")
    ) {
      location.href = appUrl("/static/index.html");
    }
    showToast("ログアウトしました");
  });
}

async function bootstrapGame() {
  if (!state.token) {
    renderAuthState();
    if (
      !qs("#authView") &&
      !(currentPage() === "mypage" && qs(".mypage-hud"))
    ) {
      location.href = appUrl("/static/index.html");
    }
    return;
  }

  renderAuthState();
  setLoading(true);
  try {
    const jobs = [];
    if (qs("#playerName") || qs("#playerLevel")) jobs.push(loadPlayer());
    if (qs("#areasList")) jobs.push(loadAreas());
    if (qs("#inventoryList")) jobs.push(loadInventory());
    if (qs("#encyclopediaList")) jobs.push(loadEncyclopedia());
    if (qs("#noticesList")) jobs.push(loadNotices());
    if (qs("#deckPreview")) jobs.push(loadDeckPreview());
    if (qs("#bossPreview")) jobs.push(loadBossPreview());
    await Promise.all(jobs);
  } finally {
    setLoading(false);
  }
}

async function init() {
  bindTabs();
  syncNav();
  bindAuthForms();
  bindGameActions();
  updateClock();
  setInterval(updateClock, 30000);
  if (await autoLoginForAndroid()) return;
  renderAuthState();
  bootstrapGame().catch(() => {});
}

init().catch((err) => showToast(err.message, "error"));
