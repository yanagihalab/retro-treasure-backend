function renderHistory() {
  const root = qs("#gachaHistory");
  if (!root) return;
  if (!state.history.length) return;
  let items = state.history.map(item => renderCardItem(item.card, !item.duplicate, item.duplicate ? "重複" : "NEW!", true));
  while(items.length < 4) items.push("<div></div>");
  root.innerHTML = items.join("");

  qsa(".card-archive-entry[data-card-id]").forEach((entryNode) => {
    entryNode.addEventListener("click", () => {
      const cardID = Number(entryNode.dataset.cardId);
      const card = state.history.find(item => item.card.id === cardID);
      openCardDialog(card.card, null);
    });
  });
}

async function drawGacha() {
  const box = qs("#gachaResultBox");
  const btn = qs("#drawGachaBtn");
  if (btn) btn.disabled = true;
  try {
    await sleep(250);
    const result = await api("/api/gacha/draw", {
      method: "POST",
      body: JSON.stringify({}),
    });
    if (result.card) {
      state.history.unshift(result);
      state.history = state.history.slice(0, 10);
      renderHistory();
    } else {
      if (box) box.textContent = "カード取得結果を表示できませんでした。";
    }
    await loadPlayer();
  } catch (err) {
    console.log(err);
    if (box) box.textContent = `ERROR: ${err.message}`;
  } finally {
    if (btn) btn.disabled = false;
  }
}

async function initGacha(player) {
  const drawBtn = qs("#drawGachaBtn");
  if (drawBtn) drawBtn.addEventListener("click", drawGacha);

  renderHistory();
}
state.history = [];
state.initializing.push(initGacha);
