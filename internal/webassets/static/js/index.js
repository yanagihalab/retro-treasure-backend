async function login(username, password) {
  setLoading(true);
  try {
    const data = await api("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
    setAuth(data.token);
    location.href = appUrl("/static/index.html");
  } finally {
    setLoading(false);
  }
}

async function registerAndLogin(username, password, tutorial = false) {
  setLoading(true);
  try {
    const data = await api("/api/auth/register", {
      method: "POST",
      body: JSON.stringify({ username, password, tutorial: tutorial, }),
    });
    setAuth(data.token);
    location.href = appUrl("/static/index.html");
  } finally {
    setLoading(false);
  }
}

async function quickStart(tutorial = false) {
  const suffix = Date.now().toString().slice(-6);
  await registerAndLogin(`player_${suffix}`, `startpass${suffix}`, tutorial);
}

async function claimBonus() {
  setLoading(true);
  try {
    const data = await api("/api/login-bonus/claim", {
      method: "POST",
      body: "{}",
    });
    showToast(`ログインボーナス獲得: ${data.reward_value} コイン`, "success");
    await loadPlayer();
  } finally {
    setLoading(false);
  }
}

async function bootstrap() {
    const ok = Boolean(state.token);
    qs("#authView")?.classList.toggle("hidden", ok);
    qs("#gameView")?.classList.toggle("hidden", !ok);

    qs("#loginForm")?.addEventListener("submit", async (event) => {
        event.preventDefault();
        const form = new FormData(event.currentTarget);
        try {
            await login(form.get("username"), form.get("password"));
        } catch (err) {
            showToast(err.message, "error");
        }
    });

    qs("#registerForm")?.addEventListener("submit", async (event) => {
        event.preventDefault();
        const form = new FormData(event.currentTarget);
        try {
            await registerAndLogin(form.get("username"), form.get("password"));
        } catch (err) {
            showToast(err.message, "error");
        }
    });

    qs("#quickStartBtn")?.addEventListener("click", async () => {
        try {
            await quickStart();
        } catch (err) {
            showToast(err.message, "error");
        }
    });

    qs("#quickStartBtn2")?.addEventListener("click", async () => {
      try {
        await quickStart(true);
      } catch (err) {
        showToast(err.message, "error");
      }
    });

    qs("#logoutBtn")?.addEventListener("click", () => {
        clearAuth();
        location.href = appUrl("/static/index.html");
    });

    qs("#claimBonusBtn")?.addEventListener("click", () =>
        claimBonus().catch((err) => showToast(err.message, "error")),
    );
}
state.initializing.push(bootstrap);