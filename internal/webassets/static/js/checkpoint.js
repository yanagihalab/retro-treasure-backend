function checkpointStatusText(entry) {
  const rec = entry.record || {};
  if ((rec.claim_count || 0) === 0) return "未到達";
  if (entry.can_claim_daily) return "本日受取可";
  return "受取済";
}

function rewardTags(cp) {
  const tags = [];
  if (
    (cp.gacha_ticket_reward || 0) > 0 ||
    cp.event_reward_type === "gacha_ticket"
  )
    tags.push("ガチャ券");
  if (
    (cp.boss_ticket_reward || 0) > 0 ||
    cp.event_reward_type === "boss_ticket"
  )
    tags.push("ボス挑戦権");
  if (cp.event_reward_type === "coin") tags.push("COIN強化");
  if (cp.event_reward_type === "exp") tags.push("EXP強化");
  if (cp.is_event_active) tags.push("イベント中");
  return tags.length ? tags : ["通常報酬"];
}

function setCheckpointAuth(token, userID, username) {
  state.token = token;
  localStorage.setItem("retro_token", token);
}

function clearCheckpointAuth() {
  state.token = "";
  localStorage.removeItem("retro_token");
}

function findHistoryEntry(cpID) {
  return state.history.find((entry) => entry.checkpoint?.id === cpID);
}

function renderSelectedCheckpointDetail(cp) {
  const code = qs("#selectedCheckpointCode");
  const root = qs("#selectedCheckpointDetail");
  if (!root) return;

  if (!cp) {
    if (code) code.textContent = "未選択";
    root.classList.add("muted");
    root.textContent =
      "マップまたは一覧からチェックポイントを選択してください。";
    return;
  }

  const entry = findHistoryEntry(cp.id);
  const rec = entry?.record || {};
  const tags = rewardTags(cp).map((tag) => `<span>${tag}</span>`).join("");
  if (code) code.textContent = cp.qr_text;
  root.classList.remove("muted");
  root.innerHTML = `
    <strong>${cp.name}（${cp.area}）</strong>
    <div class="meta">${cp.description}</div>
    <div class="checkpoint-tag-row">${tags}</div>
    <div class="checkpoint-reward-grid">
      <div><span>初回COIN</span><strong>${cp.first_reward_coin}</strong></div>
      <div><span>初回EXP</span><strong>${cp.first_reward_exp}</strong></div>
      <div><span>日次COIN</span><strong>${cp.daily_reward_coin}</strong></div>
      <div><span>日次EXP</span><strong>${cp.daily_reward_exp}</strong></div>
    </div>
    <div class="meta">到達回数: ${rec.claim_count ?? 0} / 状態: ${entry ? checkpointStatusText(entry) : "未到達"}</div>
    <div class="meta">イベント: ${cp.event_reward_name || "なし"} ${cp.event_reward_value ? `+${cp.event_reward_value}` : ""}</div>
  `;
}

function renderRouteSummary() {
  const summary = qs("#checkpointMapSummary");
  if (summary) {
    const areas = new Set(state.master.map((cp) => cp.area));
    summary.textContent = `${state.master.length}地点 / ${areas.size}エリア`;
  }
}

function createCheckpointDivIcon(qrText, selected) {
  return L.divIcon({
    className: "checkpoint-leaflet-icon-wrapper",
    html: `
      <div class="checkpoint-pin ${selected ? "selected" : ""}">
        <span class="checkpoint-pin-head"></span>
        <span class="checkpoint-pin-dot"></span>
        <span class="checkpoint-pin-label">${qrText}</span>
      </div>
    `,
    iconSize: [58, 72],
    iconAnchor: [29, 66],
    popupAnchor: [0, -58],
  });
}

function fitMapToCheckpoints() {
  if (!state.map || !state.master.length) return;

  const valid = state.master.filter(
    (cp) => typeof cp.lat === "number" && typeof cp.lng === "number",
  );

  if (!valid.length) return;

  if (valid.length === 1) {
    state.map.setView([valid[0].lat, valid[0].lng], 15);
    return;
  }

  const bounds = L.latLngBounds(valid.map((cp) => [cp.lat, cp.lng]));
  state.map.fitBounds(bounds, { padding: [30, 30] });
}

function highlightSelectedMarker(openPopup = true) {
  Object.entries(state.mapMarkers).forEach(([id, marker]) => {
    const isSelected = id === state.selectedCheckpointId;
    marker.setIcon(createCheckpointDivIcon(marker.__qrText || "", isSelected));

    if (isSelected && openPopup) {
      marker.openPopup();
    }
  });
}

function selectCheckpoint(cp, openPopup = true) {
  state.selectedCheckpointId = cp.id;
  const input = qs("#qrTextInput");
  if (input) input.value = cp.qr_text;

  renderMaster({ checkpoints: state.master });
  renderRouteSummary();
  renderSelectedCheckpointDetail(cp);
  highlightSelectedMarker(openPopup);
}

function checkpointPointBounds(checkpoints) {
  const valid = checkpoints.filter(
    (cp) => typeof cp.lat === "number" && typeof cp.lng === "number",
  );
  if (!valid.length) return null;

  const lats = valid.map((cp) => cp.lat);
  const lngs = valid.map((cp) => cp.lng);
  return {
    minLat: Math.min(...lats),
    maxLat: Math.max(...lats),
    minLng: Math.min(...lngs),
    maxLng: Math.max(...lngs),
  };
}

function fallbackPointStyle(cp, bounds, index = 0) {
  const routePositions = [
    [40, 23],
    [57, 28],
    [74, 31],
    [38, 32],
    [28, 29],
    [50, 33],
    [35, 18],
    [23, 22],
    [45, 26],
    [42, 27],
    [40, 30],
    [41, 35],
    [47, 24],
    [59, 29],
    [58, 30],
    [51, 35],
    [49, 38],
    [24, 19],
    [22, 21],
    [25, 20],
    [20, 19],
    [18, 18],
    [43, 18],
    [61, 33],
  ];
  if (routePositions[index]) {
    const [x, y] = routePositions[index];
    return `left:${x}%;top:${y}%;`;
  }

  if (!bounds) {
    return `left:${50 + (cp.map_x || 0)}%;top:${50 + (cp.map_y || 0)}%;`;
  }

  const latRange = Math.max(0.000001, bounds.maxLat - bounds.minLat);
  const lngRange = Math.max(0.000001, bounds.maxLng - bounds.minLng);
  const x = 20 + ((cp.lng - bounds.minLng) / lngRange) * 60;
  const y = 28 + (1 - (cp.lat - bounds.minLat) / latRange) * 24;
  return `left:${Math.max(20, Math.min(80, x))}%;top:${Math.max(28, Math.min(52, y))}%;`;
}

function renderFallbackMap() {
  const mapEl = qs("#checkpointMap");
  const legend = qs("#checkpointMapLegend");
  if (!mapEl) return;

  if (state.map?.remove) {
    state.map.remove();
  }
  state.map = null;
  state.mapMarkers = {};
  const bounds = checkpointPointBounds(state.master);
  const active =
    state.master.find((cp) => cp.id === state.selectedCheckpointId) ||
    state.master[0];

  mapEl.classList.add("checkpoint-fallback-map");
  mapEl.innerHTML = `
    <div class="checkpoint-map-grid" aria-hidden="true"></div>
    <div class="checkpoint-map-route" aria-hidden="true"></div>
    <div class="checkpoint-map-area-label">SUWA RELIC ROUTE</div>
    ${state.master
      .map(
        (cp, index) => `
          <button class="checkpoint-map-node ${cp.id === state.selectedCheckpointId ? "selected" : ""}" data-id="${cp.id}" type="button" style="${fallbackPointStyle(cp, bounds, index)}">
            <span>${cp.qr_text}</span>
          </button>
        `,
      )
      .join("")}
  `;

  if (legend) {
    legend.textContent = active
      ? `${active.name} / ${active.area} / ${active.qr_text}`
      : "チェックポイントMAPを表示中";
  }

  mapEl.querySelectorAll(".checkpoint-map-node").forEach((node) => {
    node.addEventListener("click", () => {
      const cp = state.master.find((x) => x.id === node.dataset.id);
      if (!cp) return;
      selectCheckpoint(cp, false);
      showToast(`${cp.name} を選択しました`, "success");
    });
  });
}

function renderMap() {
  const mapEl = qs("#checkpointMap");
  if (!mapEl) return;

  if (typeof L === "undefined") {
    renderFallbackMap();
    return;
  }

  mapEl.classList.remove("checkpoint-fallback-map");

  if (!state.map) {
    state.map = L.map("checkpointMap", {
      zoomControl: true,
      scrollWheelZoom: true,
    }).setView([36.0205, 138.13], 13);

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: "&copy; OpenStreetMap contributors",
      detectRetina: true,
      maxZoom: 19,
    }).addTo(state.map);
  }

  Object.values(state.mapMarkers).forEach((marker) => {
    state.map.removeLayer(marker);
  });
  state.mapMarkers = {};

  state.master.forEach((cp) => {
    if (typeof cp.lat !== "number" || typeof cp.lng !== "number") return;

    const marker = L.marker([cp.lat, cp.lng], {
      icon: createCheckpointDivIcon(
        cp.qr_text,
        cp.id === state.selectedCheckpointId,
      ),
    }).addTo(state.map);

    marker.__qrText = cp.qr_text;

    marker.bindPopup(`
      <div class="checkpoint-popup-title">${cp.name}</div>
      <div class="checkpoint-popup-meta">${cp.area}</div>
      <div class="checkpoint-popup-meta">入力: ${cp.qr_text}</div>
      <div class="checkpoint-popup-meta">${cp.description || ""}</div>
    `);

    marker.on("click", () => {
      selectCheckpoint(cp, true);
      showToast(`${cp.name} を選択しました`, "success");
    });

    state.mapMarkers[cp.id] = marker;
  });

  fitMapToCheckpoints();
  highlightSelectedMarker(false);

  setTimeout(() => {
    if (state.map) state.map.invalidateSize();
  }, 50);
}

function renderMaster(data) {
  state.master = data.checkpoints || [];
  if (!state.selectedCheckpointId && state.master.length) {
    state.selectedCheckpointId = state.master[0].id;
    const input = qs("#qrTextInput");
    if (input) input.value = state.master[0].qr_text;
  }
  const root = qs("#checkpointMasterList");
  if (!root) return;

  root.innerHTML = state.master
    .map((cp) => {
      const selected = cp.id === state.selectedCheckpointId;
      return `
        <button class="checkpoint-list-item ${selected ? "selected" : ""}" data-id="${cp.id}">
          <strong>${cp.name}</strong>
          <span class="checkpoint-list-area">${cp.area}</span>
          <span class="checkpoint-list-reward">COIN ${cp.first_reward_coin} / EXP ${cp.first_reward_exp}</span>
          <div class="checkpoint-tag-row"><span>${cp.qr_text}</span>${rewardTags(cp).map((tag) => `<span>${tag}</span>`).join("")}</div>
        </button>
      `;
    })
    .join("");

  root.querySelectorAll(".checkpoint-list-item").forEach((btn) => {
    btn.addEventListener("click", () => {
      const cp = state.master.find((x) => x.id === btn.dataset.id);
      if (!cp) return;
      selectCheckpoint(cp, true);
      showToast(`${cp.name} を選択しました`, "success");
    });
  });

  renderMap();
  renderRouteSummary();
  renderSelectedCheckpointDetail(
    state.master.find((cp) => cp.id === state.selectedCheckpointId),
  );
}

function renderHistory(data) {
  state.history = data.entries || [];

  setText("#gachaTicketCount", data.gacha_tickets ?? 0);
  setText("#bossTicketCount", data.boss_challenge_tickets ?? 0);
  setText("#cpStatusMini", `地点 ${state.history.length}`);
  setText(
    "#visitedCheckpointCount",
    state.history.filter((entry) => (entry.record?.claim_count || 0) > 0)
      .length,
  );
  setText(
    "#dailyReadyCount",
    state.history.filter((entry) => entry.can_claim_daily).length,
  );
  renderRouteSummary();
  renderSelectedCheckpointDetail(
    state.master.find((cp) => cp.id === state.selectedCheckpointId),
  );

  const root = qs("#checkpointHistoryList");
  if (!root) return;

  root.innerHTML = state.history
    .map((entry) => {
      const rec = entry.record || {};
      return `
        <div class="list-item">
          <strong>${entry.checkpoint.name}</strong>
          <div class="meta">状態: ${checkpointStatusText(entry)}</div>
          <div class="meta">到達回数: ${rec.claim_count ?? 0}</div>
          <div class="meta">初回: ${rec.first_claimed_at || "-"}</div>
          <div class="meta">最終: ${rec.last_claimed_at || "-"}</div>
          <div class="meta">日次受取可: ${entry.can_claim_daily ? "YES" : "NO"}</div>
          <div class="meta">イベント受取済: ${rec.event_claimed ? "YES" : "NO"}</div>
        </div>
      `;
    })
    .join("");
}

async function reloadAll() {
  const master = await api("/api/checkpoints/master");
  renderMaster(master);

  if (!state.token) {
    setText("#cpStatusMini", `地点 ${state.master.length}`);
    return;
  }

  try {
    const history = await api("/api/checkpoints/history");
    renderHistory(history);
  } catch (err) {
    if (/invalid token|missing bearer token|unauthorized/i.test(err.message)) {
      clearCheckpointAuth();
      setText("#cpStatusMini", `地点 ${state.master.length}`);
      const resultBox = qs("#checkpointResult");
      if (resultBox) resultBox.textContent = "ログインすると報酬受取と到達履歴を利用できます。";
      return;
    }
    throw err;
  }
}

async function claimCheckpoint() {
  const input = qs("#qrTextInput");
  const resultBox = qs("#checkpointResult");
  const qrText = (input?.value || "").trim();

  if (!qrText) {
    showToast("QR1〜QR24 を入力してください", "error");
    return;
  }

  if (resultBox) resultBox.textContent = "通信中...";

  try {
    const res = await api("/api/checkpoints/claim", {
      method: "POST",
      body: JSON.stringify({ qr_text: qrText }),
    });

    if (resultBox) {
      const rewards = (res.rewards || [])
        .map((r) => `<li>${r.label}: ${r.type} +${r.value}</li>`)
        .join("");

      resultBox.innerHTML = `
        <h3>${res.checkpoint.name}</h3>
        <p>${res.summary}</p>
        <ul>${rewards || "<li>受け取れる報酬はありません。</li>"}</ul>
        <p>累計到達回数: ${res.claim_count}</p>
      `;
    }

    state.selectedCheckpointId = res.checkpoint.id;
    await loadPlayer();
    await reloadAll();
    showToast("チェックポイント報酬を反映しました", "success");
  } catch (err) {
    if (resultBox) resultBox.textContent = `ERROR: ${err.message}`;
    showToast(err.message, "error");
  }
}

async function bootstrap() {
  window.scrollTo(0, 0);

  document.addEventListener("click", (event) => {
    const target = event.target;
    if (!(target instanceof Element)) return;

    if (target.closest("#claimCheckpointBtn")) {
      event.preventDefault();
      claimCheckpoint();
    }
  });

  try {
    await reloadAll();
  } catch (err) {
    console.log(err);
    const resultBox = qs("#checkpointResult");
    if (resultBox) resultBox.textContent = `ERROR: ${err.message}`;
  }
}
state.master = [];
state.history = [];
state.selectedCheckpointId = "";
state.map = null;
state.mapMarkers = {};
state.initializing.push(bootstrap);