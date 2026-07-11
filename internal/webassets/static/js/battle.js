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
  boss: null,
  deck: [],
  battleRunning: false,
  bossID: Number(new URLSearchParams(location.search).get("boss_id") || "1"),
};
const AFTER_BATTLE_REDIRECT_DELAY_MS = 2200;
const qs = (s) => document.querySelector(s);
const qsa = (s) => [...document.querySelectorAll(s)];
function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}
function setText(sel, value) {
  const el = qs(sel);
  if (!el) {
    console.warn(`element not found: ${sel}`);
    return;
  }
  el.textContent = value;
}
function showToast(message, type = "") {
  const t = qs("#toast");
  if (!t) return;
  t.textContent = message;
  t.className = `toast ${type}`.trim();
  t.classList.remove("hidden");
  clearTimeout(showToast.timer);
  showToast.timer = setTimeout(() => t.classList.add("hidden"), 2200);
}
function getAuthHeaders() {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${state.token}`,
  };
}
async function api(path, opt = {}) {
  const res = await fetch(appUrl(path), {
    ...opt,
    headers: { ...(opt.headers || {}), ...getAuthHeaders() },
  });
  const ct = res.headers.get("content-type") || "";
  const data = ct.includes("application/json")
    ? await res.json()
    : await res.text();
  if (!res.ok) throw new Error(data?.error || `request failed: ${res.status}`);
  return data;
}
function updateClock() {
  const clock = qs("#clockText");
  if (clock)
    clock.textContent = new Date().toLocaleTimeString("ja-JP", {
      hour: "2-digit",
      minute: "2-digit",
    });
}
function initials(name) {
  if (!name) return "??";
  const s = String(name).trim();
  return s.slice(0, 2).toUpperCase();
}
function cardImageUrl(id) {
  return staticUrl(`/static/img/cards/card-${String(id).padStart(2, "0")}.png?v=relic-button-icons-fit-20260702`);
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
function elementLabel(element) {
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
    }[element || "none"] ||
    element ||
    "無"
  );
}
function normalizeCardEntry(entry) {
  const base = entry?.card ? entry.card : entry || {};
  return {
    id: base.id ?? entry?.id ?? null,
    name: base.name ?? entry?.name ?? "UNKNOWN",
    rarity: base.rarity ?? entry?.rarity ?? 1,
    element: base.element ?? entry?.element ?? "none",
    max_hp: Number(base.max_hp ?? entry?.max_hp ?? entry?.hp ?? 1),
    current_hp: Number(
      entry?.current_hp ??
        base.current_hp ??
        base.max_hp ??
        entry?.max_hp ??
        entry?.hp ??
        1,
    ),
    attack: Number(base.attack ?? entry?.attack ?? 0),
    defense: Number(base.defense ?? entry?.defense ?? 0),
    slot: entry?.deck_slot ?? entry?.slot ?? null,
  };
}
function setBattleButtonsDisabled(disabled) {
  const el = qs("#startAutoBattleBtn");
  if (el) el.disabled = disabled;
}
function setBattleStatus(text) {
  setText("#battleStatusText", text);
}
function clearBattleLog() {
  const log = qs("#battleLog");
  if (log) log.innerHTML = "";
}
function escapeBattleText(value) {
  return String(value ?? "").replace(
    /[&<>"']/g,
    (ch) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        ch
      ],
  );
}
function appendBattleLog(text, extra = "", source = "system") {
  const log = qs("#battleLog");
  if (!log) return;
  const div = document.createElement("div");
  div.className = `list-item battle-log-entry source-${source} ${extra}`.trim();
  const labels = {
    boss: "BOSS SKILL",
    ally: "ALLY SKILL",
    turn: "TURN",
    system: "SYSTEM",
  };
  div.innerHTML = `<span class="battle-log-source">${labels[source] || labels.system}</span><span class="battle-log-message">${escapeBattleText(text)}</span>`;
  log.appendChild(div);
  log.scrollTop = log.scrollHeight;
}
function updateBossHP(current, max) {
  const safeMax = Math.max(1, Number(max || 1));
  const safeCurrent = Math.max(0, Number(current || 0));
  const pct = Math.max(0, Math.min(100, (safeCurrent / safeMax) * 100));
  setText("#bossHpText", `${safeCurrent} / ${safeMax}`);
  const bar = qs("#bossHpBar");
  if (bar) bar.style.width = `${pct}%`;
}
async function animateBossHP(from, to, max, duration = 320) {
  const start = Math.max(0, Number(from));
  const end = Math.max(0, Number(to));
  const diff = end - start;
  const st = performance.now();
  return new Promise((resolve) => {
    function step(now) {
      const t = Math.min(1, (now - st) / duration);
      const eased = 1 - Math.pow(1 - t, 3);
      updateBossHP(Math.round(start + diff * eased), max);
      if (t < 1) requestAnimationFrame(step);
      else {
        updateBossHP(end, max);
        resolve();
      }
    }
    requestAnimationFrame(step);
  });
}
function hpGaugeClass(pct) {
  if (pct < 20) return "hp-red";
  if (pct < 50) return "hp-yellow";
  return "hp-green";
}
function updateDeckCardHP(slot, current, max) {
  const card = qs(`#deck-card-${slot}`);
  if (!card) return;
  const safeCurrent = Math.max(0, Number(current));
  const safeMax = Math.max(1, Number(max));
  const pct = Math.max(0, Math.min(100, (safeCurrent / safeMax) * 100));
  const bar = card.querySelector(".hp-fill");
  if (bar) {
    bar.style.width = `${pct}%`;
    bar.classList.remove("hp-green", "hp-yellow", "hp-red");
    bar.classList.add(hpGaugeClass(pct));
  }
  card.setAttribute(
    "aria-label",
    `${state.deck[slot]?.name || "カード"} HP ${safeCurrent} / ${max}`,
  );
  card.classList.toggle("dead-card", safeCurrent <= 0);
}
async function animateCardHP(slot, from, to, max, duration = 320) {
  const start = Math.max(0, Number(from));
  const end = Math.max(0, Number(to));
  const diff = end - start;
  const st = performance.now();
  return new Promise((resolve) => {
    function step(now) {
      const t = Math.min(1, (now - st) / duration);
      const eased = 1 - Math.pow(1 - t, 3);
      updateDeckCardHP(slot, Math.round(start + diff * eased), max);
      if (t < 1) requestAnimationFrame(step);
      else {
        updateDeckCardHP(slot, end, max);
        resolve();
      }
    }
    requestAnimationFrame(step);
  });
}
function flashBossBar() {
  const bar = qs("#bossHpBar");
  if (!bar) return;
  bar.classList.add("damage-flash");
  setTimeout(() => bar.classList.remove("damage-flash"), 180);
}
function flashBossUnit() {
  const el = qs("#bossUnit");
  if (!el) return;
  el.classList.add("hit");
  setTimeout(() => el.classList.remove("hit"), 180);
}
function flashDeckCard(slot) {
  const el = qs(`#deck-card-${slot}`);
  if (!el) return;
  el.classList.add("damage-flash");
  setTimeout(() => el.classList.remove("damage-flash"), 180);
}
function flashAttackCard(slot) {
  const el = qs(`#deck-card-${slot}`);
  if (!el) return;
  el.classList.add("attack-flash");
  setTimeout(() => el.classList.remove("attack-flash"), 300);
}
function defenseStatusLabel(effect) {
  return (
    {
      shield: "盾",
      heal: "癒",
      mitigate: "軽",
      evade: "避",
      revive: "蘇",
      advantage: "特",
    }[effect] || "守"
  );
}
function defenseStatusText(effect) {
  return (
    {
      shield: "盾役発動",
      heal: "回復補助",
      mitigate: "軽減発動",
      evade: "回避発動",
      revive: "蘇生済み",
      advantage: "特攻防御",
    }[effect] || "防御発動"
  );
}
function addCardStatusIcon(slot, effect) {
  if (slot === null || slot === undefined || Number.isNaN(slot)) return;
  const card = qs(`#deck-card-${slot}`);
  if (!card) return;
  let tray = card.querySelector(".battle-status-icons");
  if (!tray) {
    tray = document.createElement("div");
    tray.className = "battle-status-icons";
    card.appendChild(tray);
  }
  const safeEffect = String(effect || "mitigate").replace(/[^a-z0-9_-]/gi, "");
  let icon = tray.querySelector(`.battle-status-icon.status-${safeEffect}`);
  if (!icon) {
    icon = document.createElement("span");
    icon.className = `battle-status-icon status-${safeEffect}`;
    icon.textContent = defenseStatusLabel(safeEffect);
    icon.title = defenseStatusText(safeEffect);
    tray.appendChild(icon);
  }
  icon.classList.remove("status-pop");
  void icon.offsetWidth;
  icon.classList.add("status-pop");
  card.setAttribute("data-last-status", defenseStatusText(safeEffect));
}
function ensureBattleEffectLayer() {
  let layer = qs("#battleEffectLayer");
  if (layer) return layer;
  const screen = qs(".battle-screen") || document.body;
  layer = document.createElement("div");
  layer.id = "battleEffectLayer";
  layer.className = "battle-effect-layer";
  layer.setAttribute("aria-hidden", "true");
  screen.appendChild(layer);
  return layer;
}
function effectLabel(effect) {
  return (
    {
      tentacle: "TENTACLE",
      abyss: "ABYSS",
      fire: "BLAZE",
      storm: "GALE",
      holy: "RADIANT",
      void: "VOID",
      quake: "QUAKE",
      venom: "VENOM",
      ice: "FROST",
      fang: "FANG",
      spike: "SPIKE",
      cosmic: "COSMIC",
      water: "TIDE",
    }[effect] || "RAID"
  );
}
function defenseEffectLabel(effect) {
  return (
    {
      shield: "SHIELD",
      heal: "HEAL",
      mitigate: "GUARD",
      evade: "EVADE",
      revive: "REVIVE",
      advantage: "ADVANTAGE",
    }[effect] || "GUARD"
  );
}
function skillGraphicHTML(kind = "none") {
  const safeKind = String(kind || "none").replace(/[^a-z0-9_-]/gi, "");
  return `<span class="skill-graphic skill-graphic-${safeKind}" aria-hidden="true"><i></i></span>`;
}
async function playBossAttackEffect(entry, slot) {
  const effect = entry.effect_type || "abyss";
  const layer = ensureBattleEffectLayer();
  const burst = document.createElement("div");
  burst.className = `boss-attack-effect effect-${effect}`;
  burst.innerHTML = `<div class="effect-ring"></div><div class="effect-core"></div><div class="effect-slash"></div><div class="effect-sigil">${effectLabel(effect)}</div>`;
  layer.appendChild(burst);
  const callout = document.createElement("div");
  callout.className = `skill-callout boss-skill-callout effect-${effect}`;
  callout.innerHTML = `<span>BOSS SKILL</span><div class="skill-callout-row">${skillGraphicHTML(effect)}<strong>${escapeBattleText(entry.skill_name || "BOSS ATTACK")}</strong></div>`;
  layer.appendChild(callout);
  document.body.classList.add("battle-screen-shake", `battle-effect-${effect}`);
  const boss = qs("#bossUnit");
  if (boss) boss.classList.add("boss-casting");
  const target =
    slot !== null && !Number.isNaN(slot) ? qs(`#deck-card-${slot}`) : null;
  if (target) target.classList.add("target-locked");
  await sleep(360);
  if (target) target.classList.add("impact-hit");
  await sleep(260);
  document.body.classList.remove(
    "battle-screen-shake",
    `battle-effect-${effect}`,
  );
  if (boss) boss.classList.remove("boss-casting");
  if (target) {
    target.classList.remove("target-locked");
    target.classList.remove("impact-hit");
  }
  await sleep(260);
  burst.remove();
  callout.remove();
}
async function playDefenseSkillEffect(entry, slot) {
  if (!entry.defense_skill_name || slot === null || Number.isNaN(slot)) return;
  const effect =
    entry.defense_effect_type ||
    (entry.defense_skill_tier === "advantage" ? "advantage" : "mitigate");
  const target = qs(`#deck-card-${slot}`);
  const layer = ensureBattleEffectLayer();
  const shield = document.createElement("div");
  shield.className = `defense-skill-effect ${entry.defense_skill_tier === "advantage" ? "advantage-defense" : "unique-defense"} defense-effect-${effect}`;
  shield.textContent = defenseEffectLabel(effect);
  layer.appendChild(shield);
  const callout = document.createElement("div");
  callout.className = `defense-callout ally-skill-callout ${entry.defense_skill_tier === "advantage" ? "advantage-defense" : "unique-defense"} defense-effect-${effect}`;
  callout.innerHTML = `<span>ALLY SKILL</span><div class="skill-callout-row">${skillGraphicHTML(effect)}<strong>${escapeBattleText(entry.defense_skill_name)}</strong></div>`;
  layer.appendChild(callout);
  if (target)
    target.classList.add(
      entry.defense_skill_tier === "advantage"
        ? "advantage-guard-active"
        : "unique-guard-active",
    );
  await sleep(560);
  if (target) {
    target.classList.remove("unique-guard-active");
    target.classList.remove("advantage-guard-active");
  }
  shield.remove();
  callout.remove();
}
function setActive(slot) {
  qsa(".deck-card").forEach((x) => x.classList.remove("active-attacker"));
  if (slot === null || slot === undefined || Number.isNaN(slot)) return;
  const el = qs(`#deck-card-${slot}`);
  if (el) el.classList.add("active-attacker");
}
function clearActive() {
  qsa(".deck-card").forEach((x) => x.classList.remove("active-attacker"));
}
function renderBoss(boss) {
  state.boss = boss;
  setText("#bossName", boss.name || "UNKNOWN BOSS");
  setText("#bossDescription", boss.description || "");
  const portrait = qs("#bossPortrait");
  if (portrait) {
    portrait.innerHTML = `<div class="portrait-label">BOSS</div><img id="bossInitials" class="boss-battle-image" src="${bossImageUrl(boss.id || state.bossID)}" alt="${boss.name || "BOSS"}" />`;
  }
  const badge = qs("#bossElementBadge");
  if (badge) {
    badge.textContent = elementLabel(boss.element || "none");
    badge.className = `element-badge element-${boss.element || "none"}`;
  }
  updateBossHP(boss.current_hp ?? boss.max_hp, boss.max_hp ?? 1);
}
function renderDeck(deck) {
  state.deck = (deck || [])
    .map(normalizeCardEntry)
    .sort((a, b) => (a.slot ?? 99) - (b.slot ?? 99));
  const root = qs("#deckList");
  if (!root) return;
  if (!state.deck.length) {
    root.innerHTML = '<div class="list-item">デッキがありません。</div>';
    return;
  }
  root.innerHTML = state.deck
    .map((card, index) => {
      const pct = Math.max(
        0,
        Math.min(
          100,
          (Number(card.current_hp || 0) /
            Math.max(1, Number(card.max_hp || 1))) *
            100,
        ),
      );
      return `<article class="deck-card battle-image-card element-${card.element || "none"}" id="deck-card-${index}" aria-label="${card.name} HP ${card.current_hp} / ${card.max_hp}"><img class="card-portrait-image battle-card-image" src="${cardImageUrl(card.id)}" alt="${card.name}" loading="lazy" /><div class="battle-status-icons" aria-hidden="true"></div><div class="hp-bar battle-card-hp" aria-hidden="true"><div class="hp-fill card-fill ${hpGaugeClass(pct)}" style="width:${pct}%"></div></div></article>`;
    })
    .join("");
}
function setStartOverlayVisible(show) {
  const overlay = qs("#startBattleOverlay");
  if (overlay) overlay.classList.toggle("hidden", !show);
}
async function showPreparingPopup() {
  const overlay = qs("#prepOverlay"),
    text = qs("#prepText");
  if (!overlay) return;
  overlay.classList.remove("hidden");
  if (text) text.textContent = "AUTO BATTLE 準備中.";
  await sleep(240);
  if (text) text.textContent = "AUTO BATTLE 準備中..";
  await sleep(240);
  if (text) text.textContent = "AUTO BATTLE 準備中...";
  await sleep(240);
  overlay.classList.add("hidden");
}
async function popupFlow(overlaySel, actions) {
  const overlay = qs(overlaySel);
  if (!overlay) return "stay";
  overlay.classList.remove("hidden");
  return await new Promise((resolve) => {
    const cleanup = () =>
      actions.forEach((action) => {
        const button = qs(action.selector);
        if (button) button.removeEventListener("click", action.handler);
      });
    actions.forEach((action) => {
      const button = qs(action.selector);
      action.handler = () => {
        cleanup();
        overlay.classList.add("hidden");
        resolve(action.value);
      };
      if (button) button.addEventListener("click", action.handler);
    });
  });
}
async function autoReturnToBoss(overlaySel) {
  const overlay = qs(overlaySel);
  if (overlay) {
    const subtitle = overlay.querySelector(".defeat-subtitle");
    const actions = overlay.querySelector(".defeat-actions");
    if (subtitle) subtitle.textContent = "ボス選択へ移動中...";
    if (actions) actions.style.display = "none";
    overlay.classList.remove("hidden");
  }
  setBattleStatus("ボス選択へ移動中...");
  await sleep(AFTER_BATTLE_REDIRECT_DELAY_MS);
  return "boss";
}
function goToAfterBattle(action) {
  const routes = {
    boss: "/static/boss.html",
    deck: "/static/cards.html",
    cards: "/static/cards.html",
    home: "/static/index.html",
  };
  const url = routes[action];
  if (url) window.location.href = appUrl(url);
}
async function loadBossInfo() {
  const data = await api(`/api/boss?id=${state.bossID}`);
  renderBoss(data.boss || data);
}
async function loadDeckInfo() {
  const data = await api("/api/cards/deck");
  renderDeck(data.cards || data.deck || data);
}
async function reloadBattleInfo() {
  setBattleStatus("情報を読み込んでいます...");
  await Promise.all([loadBossInfo(), loadDeckInfo()]);
  setBattleStatus("待機中...");
}
async function startAutoBattleRequest() {
  return await api("/api/boss/auto", {
    method: "POST",
    body: JSON.stringify({ boss_id: state.bossID }),
  });
}
function extractDamage(entry) {
  if (typeof entry?.damage !== "undefined") return Number(entry.damage || 0);
  const m = String(entry?.message || "").match(/に\s*(\d+)\s*ダメージ/);
  return m ? Number(m[1]) : 0;
}
function isPlayerAttackEntry(entry) {
  if (!entry) return false;
  if (entry.type === "attack") return true;
  const msg = String(entry.message || "");
  const bossName = String(state.boss?.name || entry.target_name || "");
  return (
    msg.includes("の攻撃！") &&
    (!bossName || msg.includes(bossName)) &&
    msg.includes("ダメージ")
  );
}
function isBossAttackEntry(entry) {
  if (!entry) return false;
  if (
    entry.type === "boss_attack" ||
    entry.actor_type === "boss" ||
    entry.skill_name
  )
    return true;
  const msg = String(entry.message || "");
  const bossName = String(state.boss?.name || entry.actor_name || "");
  return (
    (!bossName || msg.includes(bossName)) &&
    msg.includes("ダメージ") &&
    !msg.includes(`${bossName} に`)
  );
}
function findSlotByMessage(message, deck) {
  const msg = String(message || "");
  for (let i = 0; i < deck.length; i++) {
    const n = String(deck[i]?.name || "");
    if (n && msg.includes(n)) return i;
  }
  return null;
}
function bossSkillLogMessage(entry) {
  return `${entry.actor_name || state.boss?.name || "BOSS"} の「${entry.skill_name || "攻撃"}」！ ${entry.target_name || "味方"} に ${Number(entry.damage || 0)} ダメージ！`;
}
function allySkillLogMessage(entry) {
  if (!entry.defense_skill_name) return "";
  const effect =
    entry.defense_effect_type ||
    (entry.defense_skill_tier === "advantage" ? "advantage" : "mitigate");
  if (entry.defense_skill_tier === "advantage")
    return `${entry.target_name || "味方"} の特攻防御「${entry.defense_skill_name}」発動。${Number(entry.damage_reduced || 0)}軽減・HP${Number(entry.heal_amount || 0)}回復。`;
  if (effect === "evade")
    return `${entry.target_name || "味方"} の回避「${entry.defense_skill_name}」発動。攻撃を無効化。`;
  if (effect === "revive" && entry.revived)
    return `${entry.target_name || "味方"} の蘇生「${entry.defense_skill_name}」発動。戦線復帰。`;
  if (effect === "heal" && entry.support_heal_amount)
    return `${entry.target_name || "味方"} の回復補助「${entry.defense_skill_name}」発動。${Number(entry.damage_reduced || 0)}軽減・${entry.support_target_name || "味方"}をHP${Number(entry.support_heal_amount || 0)}回復。`;
  if (effect === "shield")
    return `${entry.target_name || "味方"} の盾役「${entry.defense_skill_name}」発動。${Number(entry.damage_reduced || 0)}軽減。`;
  return `${entry.target_name || "味方"} の軽減「${entry.defense_skill_name}」発動。${Number(entry.damage_reduced || 0)}軽減。`;
}
async function playBattleLogs(result) {
  const logs = result.logs || [];
  const bossMax = Number(result.boss?.max_hp || state.boss?.max_hp || 1);
  const deckSnapshot = (result.initial_deck || state.deck || []).map(
    normalizeCardEntry,
  );
  state.deck = deckSnapshot;
  let currentBossHP = Number(
    result.boss?.current_hp ?? result.boss?.hp ?? bossMax,
  );
  updateBossHP(currentBossHP, bossMax);
  deckSnapshot.forEach((c, i) => {
    if (typeof c.current_hp === "undefined") c.current_hp = c.max_hp;
    updateDeckCardHP(i, c.current_hp, c.max_hp);
  });
  for (const entry of logs) {
    if (entry.type === "text") {
      appendBattleLog(entry.message, "", "system");
      await sleep(380);
      continue;
    }
    if (entry.action === "turn_start") {
      setBattleStatus(
        `耐久 ${entry.round} / ${result.target_turns || 5} ターン`,
      );
      appendBattleLog(entry.message, "", "turn");
      await sleep(420);
      continue;
    }
    if (isPlayerAttackEntry(entry)) {
      const slot =
        typeof entry.card_slot !== "undefined"
          ? Number(entry.card_slot)
          : findSlotByMessage(entry.message, deckSnapshot);
      if (slot !== null && !Number.isNaN(slot)) {
        setActive(slot);
        flashAttackCard(slot);
      }
      await sleep(120);
      appendBattleLog(entry.message, "", "ally");
      await sleep(160);
      const nextBossHP =
        typeof entry.boss_hp !== "undefined"
          ? Number(entry.boss_hp)
          : Math.max(0, currentBossHP - extractDamage(entry));
      if (nextBossHP !== currentBossHP) {
        flashBossBar();
        flashBossUnit();
        await animateBossHP(currentBossHP, nextBossHP, bossMax, 320);
        currentBossHP = nextBossHP;
      }
      await sleep(300);
      clearActive();
      continue;
    }
    if (isBossAttackEntry(entry)) {
      const slot =
        typeof entry.target_slot !== "undefined"
          ? Number(entry.target_slot)
          : findSlotByMessage(entry.message, deckSnapshot);
      setBattleStatus(
        entry.skill_name ? `BOSS: ${entry.skill_name}` : "ボス攻撃",
      );
      appendBattleLog(bossSkillLogMessage(entry), "boss-skill-log", "boss");
      await playBossAttackEffect(entry, slot);
      if (entry.defense_skill_name) {
        setBattleStatus(`ALLY: ${entry.defense_skill_name}`);
        appendBattleLog(
          allySkillLogMessage(entry),
          entry.defense_skill_tier === "advantage"
            ? "rare-flash ally-skill-log"
            : "ally-skill-log",
          "ally",
        );
        addCardStatusIcon(
          slot,
          entry.defense_effect_type ||
            (entry.defense_skill_tier === "advantage"
              ? "advantage"
              : "mitigate"),
        );
        if (entry.revived) addCardStatusIcon(slot, "revive");
        if (entry.evaded) addCardStatusIcon(slot, "evade");
        await playDefenseSkillEffect(entry, slot);
      }
      if (slot !== null && !Number.isNaN(slot)) {
        const card = deckSnapshot[slot];
        const from = Number(card.current_hp ?? card.max_hp);
        const to =
          typeof entry.card_hp_after !== "undefined"
            ? Number(entry.card_hp_after)
            : Math.max(0, from - extractDamage(entry));
        flashDeckCard(slot);
        setActive(slot);
        await animateCardHP(slot, from, to, card.max_hp, 320);
        card.current_hp = to;
        state.deck[slot].current_hp = to;
      }
      if (
        typeof entry.support_target_slot !== "undefined" &&
        Number(entry.support_heal_amount || 0) > 0
      ) {
        const supportSlot = Number(entry.support_target_slot);
        if (
          !Number.isNaN(supportSlot) &&
          supportSlot !== slot &&
          deckSnapshot[supportSlot]
        ) {
          const supportCard = deckSnapshot[supportSlot];
          const from = Number(supportCard.current_hp ?? supportCard.max_hp);
          const to = Math.min(
            Number(supportCard.max_hp),
            from + Number(entry.support_heal_amount || 0),
          );
          flashDeckCard(supportSlot);
          addCardStatusIcon(supportSlot, "heal");
          await animateCardHP(supportSlot, from, to, supportCard.max_hp, 280);
          supportCard.current_hp = to;
          state.deck[supportSlot].current_hp = to;
        }
      }
      await sleep(180);
      clearActive();
      continue;
    }
    appendBattleLog(entry.message || JSON.stringify(entry), "", "system");
    await sleep(320);
  }
}
function buildLoseTestResult() {
  const initialDeck = (state.deck || []).map((c) => ({
    ...c,
    current_hp: c.current_hp ?? c.max_hp,
  }));
  const working = initialDeck.map((c) => ({
    ...c,
    current_hp: c.current_hp ?? c.max_hp,
  }));
  const boss = {
    ...(state.boss || {
      name: "遺跡の番人ゴーレム",
      max_hp: 120,
      current_hp: 120,
      element: "earth",
    }),
    current_hp: Number(state.boss?.max_hp ?? state.boss?.current_hp ?? 120),
    max_hp: Number(state.boss?.max_hp ?? 120),
  };
  let bossHP = boss.current_hp;
  const logs = [];
  for (let i = 0; i < working.length; i++) {
    const card = working[i];
    if ((card.current_hp ?? 0) <= 0) continue;
    const pd = 3;
    bossHP = Math.max(0, bossHP - pd);
    logs.push({
      type: "attack",
      card_slot: i,
      damage: pd,
      boss_hp: bossHP,
      message: `${card.name} の攻撃！ ${boss.name} に ${pd} ダメージ！`,
    });
    const bd = (card.current_hp ?? card.max_hp) + 20;
    logs.push({
      type: "boss_attack",
      target_slot: i,
      damage: bd,
      card_hp_after: 0,
      message: `${boss.name} の強撃！ ${card.name} に ${bd} ダメージ！`,
    });
    card.current_hp = 0;
  }
  logs.push({ type: "text", message: "デッキは全滅した..." });
  return {
    result: "lose",
    boss: { ...boss, current_hp: bossHP },
    initial_deck: initialDeck,
    logs,
    reward: { exp: 0, coins: 0 },
  };
}
async function renderBattleOutcome(result) {
  if (result.result === "win") {
    const turns = result.target_turns || result.survived_turns || 5;
    setBattleStatus(`耐久成功 ${turns} / ${turns}`);
    appendBattleLog(
      result.summary ||
        `耐久成功！ EXP ${result.reward?.exp || 0} / COIN ${result.reward?.coins || 0} を獲得！`,
      "rare-flash",
    );
    if (result.reward?.boss_drop_message) {
      appendBattleLog(
        result.reward.boss_drop_message,
        result.reward.boss_drop_occurred ? "rare-flash" : "",
      );
    }
    const rewardCard = result.reward?.reward_card;
    if (rewardCard) {
      appendBattleLog(`新カード獲得: ${rewardCard.name}`, "rare-flash");
    }
    return await autoReturnToBoss("#victoryOverlay");
  }
  if (result.result === "lose") {
    setBattleStatus(
      `耐久失敗 ${result.survived_turns || 0} / ${result.target_turns || 5}`,
    );
    appendBattleLog(result.summary || "デッキは全滅した...", "dead-card");
    await sleep(250);
    return await autoReturnToBoss("#defeatOverlay");
  }
  setBattleStatus("戦闘終了");
  return await autoReturnToBoss("#victoryOverlay");
}
async function runBattle(factory) {
  if (state.battleRunning) return;
  state.battleRunning = true;
  setBattleButtonsDisabled(true);
  setStartOverlayVisible(false);
  try {
    clearBattleLog();
    setBattleStatus("戦闘準備中...");
    await showPreparingPopup();
    const result = typeof factory === "function" ? await factory() : factory;
    if (result.boss) renderBoss(result.boss);
    if (result.initial_deck) renderDeck(result.initial_deck);
    await playBattleLogs(result);
    const action = await renderBattleOutcome(result);
    if (action === "retry") {
      state.battleRunning = false;
      setBattleButtonsDisabled(false);
      return runBattle(factory);
    }
    if (action && action !== "stay") {
      goToAfterBattle(action);
      return;
    }
    await Promise.all([loadBossInfo(), loadDeckInfo()]);
    setStartOverlayVisible(true);
  } catch (err) {
    console.error(err);
    appendBattleLog(`ERROR: ${err.message}`);
    showToast(err.message, "error");
    setStartOverlayVisible(true);
  } finally {
    state.battleRunning = false;
    setBattleButtonsDisabled(false);
  }
}
function bindEvents() {
  const auto = qs("#startAutoBattleBtn");
  if (auto)
    auto.addEventListener("click", () => runBattle(startAutoBattleRequest));
}
updateClock();
setInterval(updateClock, 30000);
bindEvents();
if (!state.token) {
  setBattleStatus("ログインしてください");
  appendBattleLog("先にトップ画面でログインしてください。");
  setStartOverlayVisible(false);
} else {
  setBattleButtonsDisabled(true);
  reloadBattleInfo()
    .catch((e) => showToast(e.message, "error"))
    .finally(() => setBattleButtonsDisabled(false));
}
