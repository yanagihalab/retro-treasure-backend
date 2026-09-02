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
function updateDeckCardHP(slot, current = null) {
  const card = qs(`#deck-card-${slot}`);
  const status = state.deck[slot];
  if (current != null) status.current_hp = current;
  if (!card || !status) return;
  const safeCurrent = Math.max(0, status.current_hp);
  const safeMax = Math.max(1, status.max_hp);
  const pct = Math.max(0, Math.min(100, (safeCurrent / safeMax) * 100));
  const bar = card.querySelector(".hp-fill");
  if (bar) {
    bar.style.width = `${pct}%`;
    bar.classList.remove("hp-green", "hp-yellow", "hp-red");
    bar.classList.add(pct < 20 ? "hp-red" : (pct < 50 ? "hp-yellow" : "hp-green"));
  }
  const text = card.querySelector(".hp-text");
  if (text && status) {
    text.innerHTML = `${status.current_hp}${status.shield > 0 ? (" (+" + status.shield + ")") : ""} / ${status.max_hp}`;
  }
  card.classList.toggle("dead-card", safeCurrent <= 0);
}
async function animateCardHP(slot, to, duration = 320) {
  const start = Math.max(0, state.deck[slot].current_hp);
  const end = Math.max(0, Number(to));
  const diff = end - start;
  const st = performance.now();
  return new Promise((resolve) => {
    function step(now) {
      const t = Math.min(1, (now - st) / duration);
      const eased = 1 - Math.pow(1 - t, 3);
      updateDeckCardHP(slot, Math.round(start + diff * eased));
      if (t < 1) requestAnimationFrame(step);
      else {
        updateDeckCardHP(slot, end);
        resolve();
      }
    }
    requestAnimationFrame(step);
  });
}
async function flashDeckCard(slot) {
  const el = qs(`#deck-card-${slot}`);
  if (!el) return;
  el.classList.add("damage-flash");
  await sleep(180);
  el.classList.remove("damage-flash")
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
    }[effect] || "？"
  );
}
function addCardStatusIcon(slot, effect) {
  if (slot === null || slot === undefined || Number.isNaN(slot)) return;

  // カードを入手
  const card = qs(`#deck-card-${slot}`);
  if (!card) return;

  // アイコントレイを生成
  let tray = card.querySelector(".battle-status-icons");
  if (!tray) {
    tray = document.createElement("div");
    tray.className = "battle-status-icons";
    card.appendChild(tray);
  }

  // アイコンを生成
  const safeEffect = String(effect || "mitigate").replace(/[^a-z0-9_-]/gi, "");
  let icon = tray.querySelector(`.battle-status-icon.status-${safeEffect}`);
  if (!icon) {
    icon = document.createElement("span");
    icon.className = `battle-status-icon status-${safeEffect}`;
    icon.textContent = defenseStatusLabel(safeEffect);
    tray.appendChild(icon);
  }

  // アニメーション
  icon.classList.remove("status-pop");
  void icon.offsetWidth;
  icon.classList.add("status-pop");
}
function removeCardStatusIcon(slot, effect) {
  if (slot === null || slot === undefined || Number.isNaN(slot)) return;

  // カードのdivを取得
  const card = qs(`#deck-card-${slot}`);
  if (!card) return;

  // アイコントレイを生成
  let tray = card.querySelector(".battle-status-icons");
  if (!tray) return;

  // アイコンを生成
  const safeEffect = String(effect || "mitigate").replace(/[^a-z0-9_-]/gi, "");
  let icon = tray.querySelector(`.battle-status-icon.status-${safeEffect}`);
  if (!icon) return;

  tray.removeChild(icon);
}
function ensureBattleEffectLayer() {
  let layer = qs("#battleEffectLayer");
  if (layer) return layer;
  const screen = qs(".battle-screen") || document.body;
  layer = document.createElement("div");
  layer.id = "battleEffectLayer";
  layer.className = "battle-effect-layer";
  screen.appendChild(layer);
  return layer;
}
async function playBossAttackEffect(skill, slots) {
  const effect = skill.sub_element ?? skill.element;
  const layer = ensureBattleEffectLayer();
  const burst = document.createElement("div");
  burst.className = `boss-attack-effect effect-${effect}`;
  burst.innerHTML = `<div class="effect-ring"></div><div class="effect-core"></div><div class="effect-slash"></div>`;
  layer.appendChild(burst);
  const callout = document.createElement("div");
  callout.className = `skill-callout boss-skill-callout effect-${effect}`;
  callout.innerHTML = `<span>BOSS SKILL</span><div class="skill-callout-row">${skillGraphicHTML(skill.element)}${skill.sub_element ? skillGraphicHTML(skill.sub_element) : ""}<strong>${escapeBattleText(skill.name || "BOSS ATTACK")}</strong></div>`;
  layer.appendChild(callout);
  document.body.classList.add("battle-screen-shake", `battle-effect-${effect}`);
  const boss = qs("#bossUnit");
  if (boss) boss.classList.add("boss-casting");
  let targets = [];
  for (const slot of slots) {
    const target = !Number.isNaN(slot) ? qs(`#deck-card-${slot+1}`) : null;
    if (target) targets.push(target);
  }
  for (const target of targets) target.classList.add("target-locked");
  await sleep(500);
  for (const target of targets) target.classList.add("impact-hit");
  await sleep(400);
  document.body.classList.remove(
    "battle-screen-shake",
    `battle-effect-${effect}`,
  );
  if (boss) boss.classList.remove("boss-casting");
  for (const target of targets) {
    target.classList.remove("target-locked");
    target.classList.remove("impact-hit");
  }
  await sleep(600);
  burst.remove();
  callout.remove();
}
async function playDefenseSkillEffect(skill, slot) {
  const effect = skill.effect_type;
  const target = qs(`#deck-card-${slot}`);
  const layer = ensureBattleEffectLayer();
  const callout = document.createElement("div");
  callout.className = `defense-callout ally-skill-callout`;
  callout.innerHTML = `<span>ALLY SKILL</span><div class="skill-callout-row">${skillGraphicHTML(effect)}<strong>${escapeBattleText(skill.name)}</strong></div>`;
  layer.appendChild(callout);
  if (target) target.classList.add("unique-guard-active");
  await sleep(560);
  if (target) target.classList.remove("unique-guard-active");
  callout.remove();
}
function renderBoss(boss) {
  state.boss = boss;
  setText("#bossName", boss.name || "UNKNOWN BOSS");
  const portrait = qs("#bossPortrait");
  if (portrait) {
    portrait.innerHTML = `<img src="${bossImageUrl(boss.id || state.bossID)}" alt="${boss.name || "BOSS"}" />`;
  }
}
function renderDeck() {
  const root = qs("#deckList");
  if (!root) return;
  root.innerHTML = Object.keys(state.deck).sort().map(slot => {
    const card = state.deck[slot];
    const pct = Math.max(0, Math.min(100, (Number(card.current_hp || 0) / Math.max(1, Number(card.max_hp || 1))) * 100));
    return `
    <article class="deck-card battle-image-card element-${card.element || "none"}" id="deck-card-${slot}">
      <img class="card-portrait-image" src="${cardImageUrl(card.id)}" alt="${card.name}" loading="lazy" />
      <div class="battle-status-icons"></div>
      <div class="hp-def element-${card.element}"><div class="card-preview-kicker">${elementLabel(card.element)}</div>${card.deffence}</div>
      <div class="hp-text">${card.current_hp}${card.shield > 0 ? (" (+" + card.shield + ")") : ""} / ${card.max_hp}</div>
      <div class="hp-bar"><div class="hp-fill hp-green" style="width:${pct}%"></div></div>
    </article>`;
  }).join("");
}

async function playBattleLogs(result) {
  if (!state.deck || !result.logs) return;
  const logs = result.logs;
  for (const entry of logs) {

    // ターン開始（システムメッセージ）
    if (entry.actor_type === "turn_start") {
      appendBattleLog(`耐久 ${entry.round + 1} / ${state.boss.attack_moves.length} ターン目。${state.boss.name} の猛攻に備えろ！`, "", "turn");
    }

    // ボスの攻撃
    if (entry.actor_type === "boss") {
      const slot = Number(entry.target_slot) || 0;
      const card = state.deck[slot];
      const skill = state.boss.attack_moves[entry.round][entry.skill_index];

      // メッセージ
      let text = `${state.boss.name} の「${skill.name}」！`;
      if (card.shield != entry.card_shield_after) text += `盾で ${card.shield - entry.card_shield_after} 軽減！`
      if (card.current_hp != entry.card_hp_after) text += `${card.name} に ${entry.damage} ダメージ！`
      else text += `${card.name} はダメージを無効化した。`
      appendBattleLog(text, "boss-skill-log", "boss");

      // エフェクト再生
      if (entry.index && entry.index.length > 0) {
        await playBossAttackEffect(skill, entry.index);
      }

      // 盾の変化
      card.shield = entry.card_shield_after;
      updateDeckCardHP(slot);
      if (entry.card_shield_after <= 0) removeCardStatusIcon(slot, "shield");
      if (!entry.card_evade_after) removeCardStatusIcon(slot, "evade");

      // HPの変化
      if (card.current_hp != entry.card_hp_after) {
        await flashDeckCard(slot);
        await animateCardHP(slot, entry.card_hp_after, 150);
      }
    }

    // スキルの発動
    if (entry.actor_type === "skill") {
      const slot = Number(entry.actor_slot) || 0;
      const card = state.deck[slot];
      const skill = card.skills[entry.skill_index];
      const supportSlot = Number(entry.target_slot) || 0;
      const support = state.deck[supportSlot];

      // メッセージ
      let text = `${card.name} の「${skill.name}」。`;
      if (support.current_hp != entry.card_hp_after) text += `${support.name}のHPを${entry.damage}回復。`
      if (support.shield != entry.card_shield_after) text += `${support.name}の盾が${entry.card_shield_after}になった。`
      appendBattleLog(text, "ally-skill-log", "ally");

      // エフェクト再生
      addCardStatusIcon(slot, "advantage"); // スキル使用済みアイコン
      if (entry.index && entry.index.length > 0) {
        await playDefenseSkillEffect(skill, slot); // エフェクト再生
      }
      await flashDeckCard(supportSlot); // 対象のエフェクト再生

      // 味方のHP回復効果の再生
      if (support.current_hp != entry.card_hp_after) {
        await animateCardHP(supportSlot, entry.card_hp_after, 150);
      }

      // 味方の軽減効果の再生
      if (support.shield != entry.card_shield_after) {
        addCardStatusIcon(supportSlot, "shield");
        support.shield = entry.card_shield_after;
        updateDeckCardHP(supportSlot);
      }
      if (entry.card_evade_after) addCardStatusIcon(supportSlot, "evade");
    }

    await sleep(500);
    // while(state.awaitBattle) await sleep(100);
    // state.awaitBattle = true;
  }
}
async function runBattle() {
  qs("#startBattleOverlay")?.classList.toggle("hidden", true);
  setHTML("#battleLog", "");
  await playBattleLogs(state.battleLog);
  if (state.battleLog.win) {
    qs("#victoryOverlay")?.classList.remove("hidden");
    setText("#resultText1", `${state.boss.name}の猛攻を耐え切った！`);
    setText("#resultText2", `${100 - state.boss.drop_rate}より大きければカード入手`);
  } else {
    qs("#defeatOverlay")?.classList.remove("hidden");
  }
}
function normalizeCardEntry(base, slot) {
  return {
    id: base.id ?? null,
    name: base.name,
    element: base.element,
    max_hp: Number(base.max_hp) ?? 1,
    current_hp: Number(base.max_hp) ?? 0,
    element: base.element,
    deffence: base.defense,
    shield: 0,
    skills: base.skills,
  };
}
async function bootstrap() {
  try {
    state.battleLog = await api("/api/boss/auto", { method: "POST", body: JSON.stringify({ boss_id: state.bossID }), });
    if (state.battleLog.boss) renderBoss(state.battleLog.boss);
    if (state.battleLog.initial_deck) {
      state.deck = {};
      Object.keys(state.battleLog.initial_deck).forEach(key => {
        const slot = Number(key);
        if (!slot) return;
        state.deck[slot] = normalizeCardEntry(state.battleLog.initial_deck[key], slot);
      });
      renderDeck();
    }
    qs("#startAutoBattleBtn")?.addEventListener("click", runBattle);
    qs("#resumeBattle")?.addEventListener("click", () => { state.awaitBattle = false; });
    qs("#defeatReturn")?.addEventListener("click", () => { window.location.href = appUrl("/static/boss.html") });
    qs("#victoryReward")?.addEventListener("click", getReward);
    qs("#victoryReturn")?.addEventListener("click", () => { window.location.href = appUrl("/static/boss.html") });
  } catch (err) {
    console.error(err);
    showToast(err.message, "error");
  }
}
async function getReward() {
  try {
    const result = await api("/api/boss/reward", { method: "POST", body: JSON.stringify({ boss_id: state.bossID }), });
    setText("#resultText3", `ダイス結果 = ${100 - result.drop_dice} : ${result.dropped ? "成功！" : "失敗…"}`);
    if (result.dropped) {
      if (!result.duplicate) {
        setText("#resultText4", `新カード「${result.reward_card.name}」を入手した`);
      } else {
        setText("#resultText4", `入手できるカードが無かったので、代わりに50コインを入手した`);
      }
    }
    setText("#resultText5", `獲得コイン：${result.coins}`);
    setText("#resultText6", `獲得経験値：${result.exp}`);
    qs("#victoryReward")?.classList.toggle("hidden", true);
    qs("#victoryReturn")?.classList.toggle("hidden", false);
  } catch (err) {
    console.error(err);
    showToast(err.message, "error");
  }
}

state.battleLog = null;
state.bossID = Number(new URLSearchParams(location.search).get("boss_id") || "1");
state.boss = null;
state.deck = null;
state.awaitBattle = true;
state.initializing.push(bootstrap);