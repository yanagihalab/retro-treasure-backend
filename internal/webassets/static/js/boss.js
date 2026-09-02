// ボスプレビュー
async function loadBossPreview() {
  const root = qs("#bossPreview");
  if (!root) return;

  state.bosses = await api("/api/boss");
  let items = state.bosses.map(item => item.boss).map(
    (boss) => `
        <button class="boss-list-card" type="button" data-boss-id="${boss.id}">
          <div class="portrait-art"><img src="${bossImageUrl(boss.id)}" alt="${boss.name}" loading="lazy" /></div>
          <strong>${boss.name}</strong>
          <div class="meta">EXP ${boss.reward_exp} / COIN ${boss.reward_coins}</div>
        </button>
      `,
  );
  while (items.length < 4) items.push("<div></div>");
  root.innerHTML = items.join("");

  root.querySelectorAll(".boss-list-card[data-boss-id]").forEach((node) => {
    node.addEventListener("click", () =>
      openBossPreview(Number(node.dataset.bossId)).catch((err) =>
        showToast(err.message, "error"),
      ),
    );
  });
}

function bossHintList(items = [], renderer = (item) => escapeHTML(item)) {
  return items.length ? items.map((item) => `<li>${renderer(item)}</li>`).join("") : "<span>解析中</span>";
}

// ボスダイアログ
async function openBossPreview(bossID) {
  const panel = openDialog();
  if (!panel) return;
  state.currentBossId = bossID;

  const detail = (state.bosses || []).find((item) => Number(item.boss.id) === Number(bossID));
  if (!detail) {
    panel.innerHTML = `<div id="bossDetailPanel" class="boss-detail-panel">ボス情報を読み込めませんでした。</div>`;
    return;
  }
  state.currentDetail = detail;
  const boss = detail.boss || {};
  const moves = boss.attack_moves || [];
  const hint = detail.strategy_hint || {};
  panel.innerHTML = `
  <div id="bossDetailPanel" class="boss-detail-panel muted">
    <div class="boss-detail-hero">
      <img src="${bossImageUrl(boss.id)}" alt="${escapeHTML(boss.name)}" loading="lazy" />
      <div class="boss-detail-hero-title">
        <h3>${escapeHTML(boss.name)}</h3>
        <p>${escapeHTML(boss.description || "")}</p>
        <button id="challengeBoss">このボスに挑戦</button>
      </div>
    </div>
    <section class="boss-strategy-grid">
      <article class="boss-strategy-card">
        <div class="boss-strategy-title">有効属性</div>
        <div class="boss-effective-elements">${(hint.effective_elements || []).map(element => skillGraphicHTML(element)).join("") || `<span>解析中</span>`}</div>
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
    <div class="boss-move-list">
      ${moves.map(moves => moves.map(move => `
        <article class="boss-move-item effect-${move.element || "none"}">
          <div class="boss-move-head">
            ${skillGraphicHTML(move.element)}
            ${move.sub_element ? skillGraphicHTML(move.sub_element) : ""}
            <strong>${escapeHTML(move.name)}</strong>
          </div>
          <p>攻撃属性 ${elementLabel(move.element)} / 副属性 ${subElementLabel(move.sub_element)} / 威力 ${move.power}${move.all ? "（全体）" : ""}</p>
        </article>
      `).join("")).join("")}
    </div>
    ${bossDropPreviewHTML(detail.drops[0])}
  </div>
  `;

  qsa(".card-archive-entry[data-card-id]").forEach((entryNode) => {
    entryNode.addEventListener("click", event => {
      event.preventDefault();
      event.cancelBubble = true;
      const cardID = Number(entryNode.dataset.cardId);
      const card = state.currentDetail.drops[0].candidates.find(item => item.card.id === cardID).card;
      state.previousBossId = state.currentBossId;
      openCardDialog(card, null);
    });
  });

  qs("#challengeBoss")?.addEventListener("click", event => {
    event.preventDefault();
    event.cancelBubble = true;
    location.href = appUrl(`/static/battle.html?boss_id=${state.currentBossId}`);
  });
}

function bossDropPreviewHTML(dropPreview) {
  if (!dropPreview) return "";
  const candidates = dropPreview.candidates || [];
  let items = candidates.map(item => renderCardItem(item.card, !item.owned, item.owned ? "所持済み" : "", true));
  let result = "";
  if (items) {
    while (items.length < 4) items.push("<div></div>");
    result = items.join("");
  } else {
    result =  `<div class="muted">ドロップはありません</div>`;
  }

  return `
    <section class="boss-drop-panel">
      <div class="boss-strategy-title">ボスドロップ <span>${Number(dropPreview.drop_rate)}%</span></div>
      <div class="boss-drop-list">${result}</div>
    </section>
  `;
}

state.previousBossId = null;
state.currentBossId = null;
state.currentDetail = null;
state.onCloseDialog.push(async () => {
  if (state.previousBossId) {
    openBossPreview(state.previousBossId);
    state.previousBossId = null;
  }
});
state.initializing.push(loadBossPreview);