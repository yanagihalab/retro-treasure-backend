const qs = (s) => document.querySelector(s);

const state = {
  token: localStorage.getItem("retro_token") || "",
  master: [],
  history: [],
  selectedCheckpointId: "",
  map: null,
  mapMarkers: {},
};

function setText(selector, value) {
  const el = qs(selector);
  if (el) el.textContent = value;
}

function showToast(message, type = "") {
  const toast = qs("#toast");
  if (!toast) return;

  toast.textContent = message;
  toast.className = `toast ${type}`.trim();
  toast.classList.remove("hidden");

  clearTimeout(showToast.timer);
  showToast.timer = setTimeout(() => {
    toast.classList.add("hidden");
  }, 2600);
}

function updateClock() {
  const now = new Date();
  setText(
    "#clockText",
    now.toLocaleTimeString("ja-JP", {
      hour: "2-digit",
      minute: "2-digit",
    })
  );
}

async function api(path, options = {}) {
  const headers = {
    "Content-Type": "application/json",
    "Authorization": `Bearer ${state.token}`,
    ...(options.headers || {}),
  };

  const response = await fetch(path, {
    ...options,
    headers,
  });

  const contentType = response.headers.get("content-type") || "";
  const data = contentType.includes("application/json")
    ? await response.json()
    : await response.text();

  if (!response.ok) {
    throw new Error(data?.error || `request failed: ${response.status}`);
  }

  return data;
}

function checkpointStatusText(entry) {
  const rec = entry.record || {};
  if ((rec.claim_count || 0) === 0) return "未到達";
  if (entry.can_claim_daily) return "本日受取可";
  return "受取済";
}

function createCheckpointDivIcon(qrText, selected) {
  return L.divIcon({
    className: "checkpoint-leaflet-icon-wrapper",
    html: `
      <div class="checkpoint-pin ${selected ? "selected" : ""}">
        <span class="checkpoint-pin-dot"></span>
        <span class="checkpoint-pin-label">${qrText}</span>
      </div>
    `,
    iconSize: [44, 44],
    iconAnchor: [22, 38],
    popupAnchor: [0, -30],
  });
}

function fitMapToCheckpoints() {
  if (!state.map || !state.master.length) return;

  const valid = state.master.filter(
    (cp) => typeof cp.lat === "number" && typeof cp.lng === "number"
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
  highlightSelectedMarker(openPopup);
}

function renderMap() {
  const mapEl = qs("#checkpointMap");
  if (!mapEl) return;

  if (typeof L === "undefined") {
    console.error("Leaflet is not loaded.");
    return;
  }

  if (!state.map) {
    state.map = L.map("checkpointMap", {
      zoomControl: true,
      scrollWheelZoom: true,
    }).setView([36.0205, 138.1300], 13);

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: "&copy; OpenStreetMap contributors",
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
      icon: createCheckpointDivIcon(cp.qr_text, cp.id === state.selectedCheckpointId),
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
  const root = qs("#checkpointMasterList");
  if (!root) return;

  root.innerHTML = state.master
    .map((cp) => {
      const selected = cp.id === state.selectedCheckpointId;
      return `
        <button class="checkpoint-list-item ${selected ? "selected" : ""}" data-id="${cp.id}">
          <strong>${cp.name}</strong>
          <div class="meta">入力: ${cp.qr_text} / エリア: ${cp.area}</div>
          <div class="meta">初回: COIN ${cp.first_reward_coin} / EXP ${cp.first_reward_exp}</div>
          <div class="meta">日次: COIN ${cp.daily_reward_coin} / EXP ${cp.daily_reward_exp}</div>
          <div class="meta">イベント: ${cp.event_reward_name || "なし"}</div>
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
}

function renderHistory(data) {
  state.history = data.entries || [];

  setText("#gachaTicketCount", data.gacha_tickets ?? 0);
  setText("#bossTicketCount", data.boss_challenge_tickets ?? 0);
  setText("#cpStatusMini", `地点 ${state.history.length}`);

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
  const [master, history] = await Promise.all([
    api("/api/checkpoints/master"),
    api("/api/checkpoints/history"),
  ]);

  renderMaster(master);
  renderHistory(history);
}

async function claimCheckpoint() {
  const input = qs("#qrTextInput");
  const resultBox = qs("#checkpointResult");
  const qrText = (input?.value || "").trim();

  if (!qrText) {
    showToast("QR1 / QR2 / QR3 を入力してください", "error");
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
    await reloadAll();
    showToast("チェックポイント報酬を反映しました", "success");
  } catch (err) {
    if (resultBox) resultBox.textContent = `ERROR: ${err.message}`;
    showToast(err.message, "error");
  }
}

function bindEvents() {
  qs("#claimCheckpointBtn")?.addEventListener("click", claimCheckpoint);

  qs("#reloadCheckpointBtn")?.addEventListener("click", () => {
    reloadAll().catch((e) => showToast(e.message, "error"));
  });
}

async function init() {
  updateClock();
  setInterval(updateClock, 30000);
  bindEvents();

  if (!state.token) {
    const resultBox = qs("#checkpointResult");
    if (resultBox) resultBox.textContent = "先にトップページでログインしてください。";
    return;
  }

  try {
    await reloadAll();
  } catch (err) {
    const resultBox = qs("#checkpointResult");
    if (resultBox) resultBox.textContent = `ERROR: ${err.message}`;
  }
}

init();