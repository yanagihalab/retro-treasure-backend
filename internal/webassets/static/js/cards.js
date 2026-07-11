const state = {
  token: localStorage.getItem("retro_token") || "",
  collection: [],
  deckCardIds: [],
};
const qs = (s) => document.querySelector(s);
function showToast(message, type = "") {
  const t = qs("#toast");
  t.textContent = message;
  t.className = `toast ${type}`.trim();
  t.classList.remove("hidden");
  clearTimeout(showToast.timer);
  showToast.timer = setTimeout(() => t.classList.add("hidden"), 2400);
}
function rarityStars(r) {
  return `<span class="rarity">${"★".repeat(Math.max(1, r || 1))}</span>`;
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
function elementBadge(el) {
  return `<span class="element-badge ${el} element-${el}">${elementLabel(el)}</span>`;
}
function cardImageUrl(id) {
  return `/static/img/cards/card-${String(id).padStart(2, "0")}.png?v=relic-button-icons-fit-20260702`;
}
async function api(path, opt = {}) {
  const headers = {
    "Content-Type": "application/json",
    ...(opt.headers || {}),
  };
  if (state.token) headers.Authorization = `Bearer ${state.token}`;
  const res = await fetch(path, { ...opt, headers });
  const ct = res.headers.get("content-type") || "";
  const data = ct.includes("application/json")
    ? await res.json()
    : await res.text();
  if (!res.ok) throw new Error(data?.error || `request failed: ${res.status}`);
  return data;
}
function updateClock() {
  qs("#clockText").textContent = new Date().toLocaleTimeString("ja-JP", {
    hour: "2-digit",
    minute: "2-digit",
  });
}
async function loadPlayer() {
  const p = await api("/api/player/me");
  qs("#playerCoins").textContent = p.coins;
}
async function loadCollection() {
  const data = await api("/api/cards/collection");
  state.collection = data.cards || [];
  const deck = await api("/api/cards/deck");
  state.deckCardIds = (deck.cards || [])
    .sort((a, b) => a.deck_slot - b.deck_slot)
    .map((c) => c.card.id);
  renderDeckEditors();
  renderCollection();
}
function renderDeckEditors() {
  const owned = state.collection.map((x) => x.card);
  const root = qs("#deckEditors");
  root.innerHTML = Array.from({ length: 6 })
    .map((_, i) => {
      const value = state.deckCardIds[i] || "";
      const options = owned
        .map(
          (c) =>
            `<option value="${c.id}" ${String(c.id) === String(value) ? "selected" : ""}>${c.name} [${elementLabel(c.element)}]</option>`,
        )
        .join("");
      return `<label class="deck-editor-slot">SLOT ${i + 1}<select data-slot="${i}"><option value="">未選択</option>${options}</select></label>`;
    })
    .join("");
  root.querySelectorAll("select").forEach((sel) =>
    sel.addEventListener("change", () => {
      state.deckCardIds[Number(sel.dataset.slot)] = Number(sel.value || 0);
    }),
  );
}
function renderCollection() {
  const root = qs("#cardCollection");
  root.innerHTML = state.collection
    .map(
      (entry) =>
        `<div class="list-item card-row"><img class="card-list-image" src="${cardImageUrl(entry.card.id)}" alt="${entry.card.name}" loading="lazy" /><div class="card-row-body"><strong>${entry.card.name}</strong> ${rarityStars(entry.card.rarity)} ${elementBadge(entry.card.element)}<div class="meta">Lv ${entry.user.level} / HP ${entry.card.max_hp} / ATK ${entry.card.attack} / DEF ${entry.card.defense}</div><div class="meta">${entry.in_deck ? `デッキ S${entry.user.deck_slot}` : "控え"}</div></div><button data-upgrade="${entry.card.id}">強化</button></div>`,
    )
    .join("");
  root.querySelectorAll("[data-upgrade]").forEach((btn) =>
    btn.addEventListener("click", async () => {
      try {
        const res = await api("/api/cards/upgrade", {
          method: "POST",
          body: JSON.stringify({ card_id: Number(btn.dataset.upgrade) }),
        });
        showToast(`${res.card.name} を強化しました`, "success");
        await loadPlayer();
        await loadCollection();
      } catch (err) {
        showToast(err.message, "error");
      }
    }),
  );
}
async function saveDeck() {
  const ids = state.deckCardIds.filter(Boolean);
  if (ids.length !== 6) {
    showToast("6枚すべて選択してください", "error");
    return;
  }
  try {
    await api("/api/cards/deck", {
      method: "POST",
      body: JSON.stringify({ card_ids: ids }),
    });
    showToast("デッキを保存しました", "success");
    await loadCollection();
  } catch (err) {
    showToast(err.message, "error");
  }
}
async function drawGacha() {
  try {
    const res = await api("/api/gacha/draw", { method: "POST", body: "{}" });
    qs("#gachaResult").innerHTML =
      `<div class="gacha-card-result"><img class="gacha-card-image" src="${cardImageUrl(res.card.id)}" alt="${res.card.name}" /><div><h3>${res.duplicate ? "重複" : "新規"}: ${res.card.name}</h3><p>${rarityStars(res.card.rarity)} ${elementBadge(res.card.element)}</p><p>HP ${res.card.max_hp} / ATK ${res.card.attack} / DEF ${res.card.defense}</p><p>${res.bonus_message || "新たな仲間を獲得しました。"}</p></div></div>`;
    showToast("ガチャを引きました", "success");
    await loadPlayer();
    await loadCollection();
  } catch (err) {
    showToast(err.message, "error");
  }
}
function bind() {
  qs("#reloadCardsBtn").addEventListener("click", () =>
    bootstrap().catch((e) => showToast(e.message, "error")),
  );
  qs("#saveDeckBtn").addEventListener("click", () => saveDeck());
  qs("#drawGachaBtn").addEventListener("click", () => drawGacha());
}
async function bootstrap() {
  if (!state.token) {
    showToast("先にログインしてください", "error");
    return;
  }
  await Promise.all([loadPlayer(), loadCollection()]);
}
updateClock();
setInterval(updateClock, 30000);
bind();
bootstrap().catch((e) => showToast(e.message, "error"));
