async function loadCollection() {
  const data = await api("/api/cards/collection");
  state.collection = data.cards || [];

  const deck = await api("/api/cards/deck");
  state.deckCardIds = (deck.cards || []).sort((a, b) => a.deck_slot - b.deck_slot).map((c) => c.card.id);

  renderCollection();
}

function renderCollection() {

  const root2 = qs("#deckPreview");
  let items2 = state.collection.filter(entry => entry.in_deck).map(entry => renderCardItem(entry.card, true, null, false, ["詳細", "除外"]));
  if (items2.length) {
    while (items2.length < 4) items2.push("<div></div>");
    root2.innerHTML = items2.join("");
  } else {
    root2.innerHTML = "該当するカードがありません。";
  }

  root2.querySelectorAll("#cardItemButton0[data-card-id]").forEach((btn) =>
    btn.addEventListener("click", async () => {
      const cardID = Number(btn.dataset.cardId);
      const card = state.collection.find(entry => entry.card?.id === cardID);
      openCardDialog(card.card, null);
    })
  );

  root2.querySelectorAll("#cardItemButton1[data-card-id]").forEach((btn) =>
    btn.addEventListener("click", async () => {
      try {
        const cardID = Number(btn.dataset.cardId);
        state.deckCardIds = state.deckCardIds.filter(id => id != cardID);
        await saveDeck();
      } catch (err) {
        showToast(err.message, "error");
      }
    })
  );


  const root = qs("#cardCollection");
  let items = state.collection.filter(entry => !entry.in_deck).map(entry => renderCardItem(entry.card, true, null, false, ["詳細", "追加"]));
  if (items.length) {
    while (items.length < 4) items.push("<div></div>");
    root.innerHTML = items.join("");
  } else {
    root.innerHTML = "該当するカードがありません。";
  }

  root.querySelectorAll("#cardItemButton0[data-card-id]").forEach((btn) =>
    btn.addEventListener("click", async () => {
      const cardID = Number(btn.dataset.cardId);
      const card = state.collection.find(entry => entry.card?.id === cardID);
      openCardDialog(card.card, null);
    })
  );

  root.querySelectorAll("#cardItemButton1[data-card-id]").forEach((btn) =>
    btn.addEventListener("click", async () => {
      try {
        const cardID = Number(btn.dataset.cardId);
        state.deckCardIds.push(cardID);
        await saveDeck();
      } catch (err) {
        showToast(err.message, "error");
      }
    })
  );
}

async function saveDeck() {
  const ids = state.deckCardIds.filter(Boolean);
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

state.initializing.push(loadCollection);