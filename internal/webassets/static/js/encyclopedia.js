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
  if (!entries.length) {
    list.innerHTML = '<div>該当するカードがありません。</div>';
    return;
  }
  let items = entries.map((entry) => {
      const status = entry.obtained ? entry.in_deck ? `デッキ${entry.deck_slot}` : "所持済み" : "未所持";
      return renderCardItem(entry.card, entry.obtained, status, true);
    });
  while(items.length < 4) items.push("<div></div>");
  list.innerHTML = items.join("");

  qsa(".card-archive-entry[data-card-id]").forEach((entryNode) => {
    entryNode.addEventListener("click", () => {
      const cardID = Number(entryNode.dataset.cardId);
      const entry = (state.cardArchiveEntries || []).find((entry) => entry.card?.id === cardID);
      const card = entry?.card || null;
      const status = entry.obtained ? entry.in_deck ? `デッキ${entry.deck_slot}` : "所持済み" : "未所持";
      openCardDialog(card, status);
    });
  });
}

async function bootstrap() {
  qsa("[data-card-filter]").forEach((button) =>
    button.addEventListener("click", () => {
      state.cardArchiveFilter = button.dataset.cardFilter || "all";
      qsa("[data-card-filter]").forEach((node) =>
        node.classList.toggle("active", node === button),
      );
      renderCardArchive();
    }),
  );

  await Promise.all([loadEncyclopedia()]);
}
state.initializing.push(bootstrap);