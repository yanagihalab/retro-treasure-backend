const qs = (s) => document.querySelector(s);
const state = { token: localStorage.getItem("retro_token") || "", history: [] };
function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
function setText(selector, value) {
  const el = qs(selector);
  if (el) el.textContent = value;
}
function initials(name) {
  if (!name) return "??";
  const parts = String(name).trim().split(/\s+/);
  if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
  return String(name).slice(0, 2).toUpperCase();
}
function cardImageUrl(id) {
  return `/static/img/cards/card-${String(id).padStart(2, "0")}.png?v=relic-boss-drops-deck-icons-20260630`;
}
function elementLabel(el) {
  return (
    {
      heart: "心",
      tech: "技",
      body: "体",
      fire: "火",
      water: "水",
      earth: "木",
      wind: "風",
      light: "光",
      dark: "闇",
      wood: "木",
      none: "無",
    }[el || "none"] ||
    el ||
    "無"
  );
}
function updateClock() {
  const now = new Date();
  setText(
    "#clockText",
    now.toLocaleTimeString("ja-JP", { hour: "2-digit", minute: "2-digit" }),
  );
}
async function api(path, options = {}) {
  const headers = {
    "Content-Type": "application/json",
    Authorization: `Bearer ${state.token}`,
    ...(options.headers || {}),
  };
  const response = await fetch(path, { ...options, headers });
  const ct = response.headers.get("content-type") || "";
  const data = ct.includes("application/json")
    ? await response.json()
    : await response.text();
  if (!response.ok)
    throw new Error(data?.error || `request failed: ${response.status}`);
  return data;
}
async function loadPlayerInfo() {
  const data = await api("/api/player/me");
  setText("#gachaCoinsText", `COIN: ${data.coins ?? 0}`);
}
function renderHistory() {
  const root = qs("#gachaHistory");
  if (!root) return;
  if (!state.history.length) {
    root.innerHTML = '<div class="list-item">結果はここに表示されます。</div>';
    return;
  }
  root.innerHTML = state.history
    .map(
      (item) =>
        `<div class="list-item card-row"><img class="card-list-image" src="${cardImageUrl(item.id)}" alt="${item.name}" loading="lazy" /><div><strong>${item.name}</strong><div class="meta">属性: ${elementLabel(item.element)} / レア: ${"★".repeat(Math.max(1, item.rarity || 1))}</div></div></div>`,
    )
    .join("");
}
async function drawGacha() {
  const box = qs("#gachaResultBox");
  const btn = qs("#drawGachaBtn");
  if (btn) btn.disabled = true;
  try {
    if (box) box.textContent = "通信中...";
    await sleep(250);
    const result = await api("/api/gacha/draw", {
      method: "POST",
      body: JSON.stringify({}),
    });
    const card =
      result.card || result.reward_card || result.result_card || null;
    if (card) {
      const stage = qs("#gachaResultInitials");
      if (stage) {
        stage.outerHTML = `<img id="gachaResultInitials" class="gacha-stage-image" src="${cardImageUrl(card.id)}" alt="${card.name}" />`;
      }
      if (box) {
        box.innerHTML = `<div class="gacha-card-result"><img class="gacha-card-image" src="${cardImageUrl(card.id)}" alt="${card.name}" /><div><h3>${card.name}</h3><p>属性: ${elementLabel(card.element)}</p><p>レア: ${"★".repeat(Math.max(1, card.rarity || 1))}</p><p>HP ${card.max_hp} / ATK ${card.attack} / DEF ${card.defense}</p></div></div>`;
      }
      state.history.unshift(card);
      state.history = state.history.slice(0, 10);
      renderHistory();
    } else {
      if (box) box.textContent = "カード取得結果を表示できませんでした。";
    }
    await loadPlayerInfo();
  } catch (err) {
    if (box) box.textContent = `ERROR: ${err.message}`;
  } finally {
    if (btn) btn.disabled = false;
  }
}
function bindEvents() {
  const drawBtn = qs("#drawGachaBtn");
  if (drawBtn) drawBtn.addEventListener("click", drawGacha);
  const reloadBtn = qs("#reloadGachaBtn");
  if (reloadBtn)
    reloadBtn.addEventListener("click", async () => {
      try {
        await loadPlayerInfo();
      } catch (err) {
        const box = qs("#gachaResultBox");
        if (box) box.textContent = `ERROR: ${err.message}`;
      }
    });
}
async function init() {
  updateClock();
  setInterval(updateClock, 30000);
  bindEvents();
  if (!state.token) {
    const box = qs("#gachaResultBox");
    if (box) box.textContent = "先にトップ画面でログインしてください。";
    return;
  }
  try {
    await loadPlayerInfo();
    renderHistory();
  } catch (err) {
    const box = qs("#gachaResultBox");
    if (box) box.textContent = `ERROR: ${err.message}`;
  }
}
init();
