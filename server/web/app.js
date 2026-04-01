const els = {
  loginView: document.getElementById("loginView"),
  appView: document.getElementById("appView"),
  languageSelect: document.getElementById("languageSelect"),
  loginForm: document.getElementById("loginForm"),
  loginTitle: document.getElementById("loginTitle"),
  loginSubtitle: document.getElementById("loginSubtitle"),
  labelUsername: document.getElementById("labelUsername"),
  labelPassword: document.getElementById("labelPassword"),
  loginBtn: document.getElementById("loginBtn"),
  appTitle: document.getElementById("appTitle"),
  username: document.getElementById("username"),
  password: document.getElementById("password"),
  loginError: document.getElementById("loginError"),
  statusText: document.getElementById("statusText"),
  transportText: document.getElementById("transportText"),
  sectionDevices: document.getElementById("sectionDevices"),
  sectionSession: document.getElementById("sectionSession"),
  labelDisplay: document.getElementById("labelDisplay"),
  labelQuality: document.getElementById("labelQuality"),
  labelFps: document.getElementById("labelFps"),
  qualityLow: document.getElementById("qualityLow"),
  qualityMedium: document.getElementById("qualityMedium"),
  qualityHigh: document.getElementById("qualityHigh"),
  qualityUltra: document.getElementById("qualityUltra"),
  placeholderTitle: document.getElementById("placeholderTitle"),
  placeholderHint: document.getElementById("placeholderHint"),
  refreshBtn: document.getElementById("refreshBtn"),
  connectBtn: document.getElementById("connectBtn"),
  disconnectBtn: document.getElementById("disconnectBtn"),
  mobileKeyboardBtn: document.getElementById("mobileKeyboardBtn"),
  mobileDragToggleBtn: document.getElementById("mobileDragToggleBtn"),
  mobileImeInput: document.getElementById("mobileImeInput"),
  deviceList: document.getElementById("deviceList"),
  displaySelect: document.getElementById("displaySelect"),
  qualitySelect: document.getElementById("qualitySelect"),
  fpsInput: document.getElementById("fpsInput"),
  fpsValue: document.getElementById("fpsValue"),
  stage: document.getElementById("stage"),
  remoteVideo: document.getElementById("remoteVideo")
};

const LANGUAGE_STORAGE_KEY = "remote_control_lang";
const I18N = {
  en: {
    loginTitle: "Remote Control",
    loginSubtitle: "Sign in and choose an online device to start a session.",
    username: "Username",
    password: "Password",
    signIn: "Sign In",
    appTitle: "WebRTC Remote Desktop",
    refreshDevices: "Refresh Devices",
    connect: "Connect",
    disconnect: "Disconnect",
    keyboard: "Keyboard",
    onlineDevices: "Online Devices",
    sessionOptions: "Session Options",
    display: "Display",
    quality: "Quality",
    fps: "FPS",
    qualityLow: "Low",
    qualityMedium: "Medium",
    qualityHigh: "High",
    qualityUltra: "Ultra",
    noActiveSession: "No active session",
    noActiveSessionHint: "This area shows placeholder content until a device stream is connected.",
    statusNotConnected: "Not connected",
    dragOn: "Drag: On",
    dragOff: "Drag: Off",
    dragModeDisabled: "Drag mode disabled",
    dragModeEnabled: "Drag mode enabled",
    softKeyboardOpened: "Soft keyboard opened. Input is sent to remote.",
    loginFailed: "Login failed",
    roleNotSupported: "This console supports viewer/admin only",
    signedInSelectDevice: "Signed in. Select a device to start.",
    runtimeConfigFailed: "Failed to load runtime config",
    websocketConnectFailed: "WebSocket connection failed",
    signalDisconnected: "Signal channel disconnected",
    errorPrefix: "Error",
    sessionCreated: "Session created, waiting for media...",
    sessionEnded: "Session ended",
    loadDevicesFailed: "Load devices failed",
    noOnlineDevices: "No online devices",
    defaultDisplay: "Default display",
    displaysSuffix: "displays",
    selectDeviceFirst: "Please select a device first",
    signalUnavailable: "Signal channel unavailable",
    sessionRequestSent: "Session request sent",
    sessionConfigApplied: "Session quality/FPS updated",
    mediaConnected: "Media connected",
    webrtcState: "WebRTC state",
    p2pFailedHint: "P2P transport failed. Configure TURN mode for better reachability.",
    transportTitle: "Transport",
    transportConfigured: "configured",
    transportActive: "active",
    transportUnknown: "unknown",
    transportIdle: "idle",
    transportDetecting: "detecting",
    transportP2PDirect: "P2P direct",
    transportTURNRelay: "TURN relay"
  },
  zh: {
    loginTitle: "远程控制",
    loginSubtitle: "登录后选择在线设备并发起会话。",
    username: "用户名",
    password: "密码",
    signIn: "登录",
    appTitle: "WebRTC 远程桌面",
    refreshDevices: "刷新设备",
    connect: "连接",
    disconnect: "断开",
    keyboard: "软键盘",
    onlineDevices: "在线设备",
    sessionOptions: "会话参数",
    display: "显示器",
    quality: "清晰度",
    fps: "帧率",
    qualityLow: "低",
    qualityMedium: "中",
    qualityHigh: "高",
    qualityUltra: "极高",
    noActiveSession: "未连接到设备",
    noActiveSessionHint: "连接设备后此区域将显示实时画面。",
    statusNotConnected: "未连接",
    dragOn: "拖拽: 开",
    dragOff: "拖拽: 关",
    dragModeDisabled: "拖拽模式已关闭",
    dragModeEnabled: "拖拽模式已开启",
    softKeyboardOpened: "已唤起软键盘，输入会发送到远端。",
    loginFailed: "登录失败",
    roleNotSupported: "该控制台仅支持 viewer/admin 角色",
    signedInSelectDevice: "登录成功，请先选择设备。",
    runtimeConfigFailed: "加载运行配置失败",
    websocketConnectFailed: "WebSocket 连接失败",
    signalDisconnected: "信令连接已断开",
    errorPrefix: "错误",
    sessionCreated: "会话已创建，正在等待媒体流...",
    sessionEnded: "会话结束",
    loadDevicesFailed: "拉取设备失败",
    noOnlineDevices: "暂无在线设备",
    defaultDisplay: "默认显示器",
    displaysSuffix: "个显示器",
    selectDeviceFirst: "请先选择设备",
    signalUnavailable: "信令连接不可用",
    sessionRequestSent: "会话请求已发送",
    sessionConfigApplied: "会话画质/帧率已更新",
    mediaConnected: "媒体流已连接",
    webrtcState: "WebRTC 状态",
    p2pFailedHint: "P2P 连接失败，建议配置 TURN 以提升可达性。",
    transportTitle: "链路",
    transportConfigured: "配置",
    transportActive: "当前",
    transportUnknown: "未知",
    transportIdle: "空闲",
    transportDetecting: "检测中",
    transportP2PDirect: "P2P 直连",
    transportTURNRelay: "TURN 中继"
  }
};

const state = {
  token: "",
  role: "",
  ws: null,
  devices: [],
  selectedDeviceId: "",
  sessionId: "",
  pc: null,

  lastMouseMoveAt: 0,
  isMobileMode: detectMobileMode(),

  mobileTouches: new Map(),
  mobileGesture: "idle", // idle | single | scroll
  mobilePrimaryPointerId: null,
  mobileLongPressTimer: null,
  mobileLongPressFired: false,
  mobileScrollCenter: null,
  mobileScrollRemainderY: 0,
  mobileDragMode: false,
  mobileDragPressed: false,
  mobileDragPointerId: null,

  imePrevValue: "",
  imePendingBackspace: 0,

  configuredTransportMode: "p2p",
  configuredForceRelay: false,
  configuredIceServers: [{ urls: ["stun:stun.l.google.com:19302"] }],
  activeTransportMode: "none", // none | unknown | p2p | turn
  connectedOnce: false,
  p2pWarningSent: false,
  p2pWatchdogTimer: null,
  transportProbeTimer: null,
  sessionUpdateTimer: null,
  language: detectInitialLanguage()
};

initLanguage();

if (state.isMobileMode) {
  document.body.classList.add("mobile-mode");
}

setStatus(t("statusNotConnected"));
els.fpsValue.textContent = els.fpsInput.value;
els.fpsInput.addEventListener("input", () => {
  els.fpsValue.textContent = els.fpsInput.value;
  scheduleSessionConfigUpdate();
});
els.fpsInput.addEventListener("change", () => {
  scheduleSessionConfigUpdate(0);
});

els.loginForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  els.loginError.textContent = "";

  const username = els.username.value.trim();
  const password = els.password.value;

  try {
    const loginRes = await fetch("/api/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password })
    });
    const loginData = await loginRes.json();
    if (!loginRes.ok) {
      throw new Error(loginData.error || t("loginFailed"));
    }

    state.token = loginData.token;
    state.role = loginData.role;

    if (state.role !== "viewer" && state.role !== "admin") {
      throw new Error(t("roleNotSupported"));
    }

    await loadRuntimeConfig();
    await connectWS();

    els.loginView.classList.add("hidden");
    els.appView.classList.remove("hidden");

    setStatus(t("signedInSelectDevice"));
    renderTransportStatus();
    await loadDevices();
  } catch (err) {
    els.loginError.textContent = err.message;
  }
});

els.refreshBtn.addEventListener("click", () => {
  loadDevices();
});

els.connectBtn.addEventListener("click", async () => {
  await startSession();
});

els.disconnectBtn.addEventListener("click", () => {
  stopSession(true);
});

els.qualitySelect.addEventListener("change", () => {
  sendSessionUpdate();
});

if (!state.isMobileMode) {
  bindDesktopControls();
} else {
  bindMobileControls();
  bindMobileGestureActions();
  updateMobileDragToggleButton();
}

bindMobileKeyboardBridge();
renderTransportStatus();

window.addEventListener("beforeunload", () => {
  try {
    if (state.ws) {
      state.ws.close();
    }
  } catch (_) {
    // noop
  }
});

function detectInitialLanguage() {
  const saved = localStorage.getItem(LANGUAGE_STORAGE_KEY);
  if (saved === "en" || saved === "zh") {
    return saved;
  }
  return navigator.language && navigator.language.toLowerCase().startsWith("zh") ? "zh" : "en";
}

function initLanguage() {
  if (els.languageSelect) {
    els.languageSelect.value = state.language;
    els.languageSelect.addEventListener("change", () => {
      const next = els.languageSelect.value === "zh" ? "zh" : "en";
      state.language = next;
      localStorage.setItem(LANGUAGE_STORAGE_KEY, next);
      applyLanguage();
      renderTransportStatus();
      updateMobileDragToggleButton();
    });
  }
  applyLanguage();
}

function applyLanguage() {
  document.documentElement.lang = state.language === "zh" ? "zh-CN" : "en";

  els.loginTitle.textContent = t("loginTitle");
  els.loginSubtitle.textContent = t("loginSubtitle");
  els.labelUsername.textContent = t("username");
  els.labelPassword.textContent = t("password");
  els.loginBtn.textContent = t("signIn");
  els.appTitle.textContent = t("appTitle");
  els.refreshBtn.textContent = t("refreshDevices");
  els.connectBtn.textContent = t("connect");
  els.disconnectBtn.textContent = t("disconnect");
  els.mobileKeyboardBtn.textContent = t("keyboard");
  els.sectionDevices.textContent = t("onlineDevices");
  els.sectionSession.textContent = t("sessionOptions");
  els.labelDisplay.textContent = t("display");
  els.labelQuality.textContent = t("quality");
  els.labelFps.textContent = t("fps");
  els.qualityLow.textContent = t("qualityLow");
  els.qualityMedium.textContent = t("qualityMedium");
  els.qualityHigh.textContent = t("qualityHigh");
  if (els.qualityUltra) {
    els.qualityUltra.textContent = t("qualityUltra");
  }
  els.placeholderTitle.textContent = t("noActiveSession");
  els.placeholderHint.textContent = t("noActiveSessionHint");

  if (state.devices.length > 0 || els.deviceList.children.length > 0) {
    renderDevices();
    populateDisplayOptions();
  }
  renderTransportStatus();
  updateMobileDragToggleButton();

  if (!isSessionActive()) {
    setStatus(t("statusNotConnected"));
  }
}

function t(key) {
  const bundle = I18N[state.language] || I18N.en;
  if (Object.prototype.hasOwnProperty.call(bundle, key)) {
    return bundle[key];
  }
  return I18N.en[key] || key;
}

function bindDesktopControls() {
  els.stage.addEventListener("click", () => {
    els.stage.focus();
  });

  els.stage.addEventListener("mousemove", (event) => {
    if (!isSessionActive()) {
      return;
    }
    const now = Date.now();
    if (now - state.lastMouseMoveAt < 30) {
      return;
    }
    state.lastMouseMoveAt = now;

    const pos = normalizedPositionFromClient(event.clientX, event.clientY);
    sendControl({ kind: "mouse_move", x: pos.x, y: pos.y });
  });

  els.stage.addEventListener("mousedown", (event) => {
    if (!isSessionActive()) {
      return;
    }
    const pos = normalizedPositionFromClient(event.clientX, event.clientY);
    sendControl({ kind: "mouse_move", x: pos.x, y: pos.y });
    sendControl({ kind: "mouse_click", button: mapMouseButton(event.button), pressed: true });
  });

  els.stage.addEventListener("mouseup", (event) => {
    if (!isSessionActive()) {
      return;
    }
    sendControl({ kind: "mouse_click", button: mapMouseButton(event.button), pressed: false });
  });

  els.stage.addEventListener("contextmenu", (event) => {
    if (isSessionActive()) {
      event.preventDefault();
    }
  });

  els.stage.addEventListener("keydown", (event) => {
    if (!isSessionActive()) {
      return;
    }
    sendControl({ kind: "key", key: event.key, pressed: true });
    event.preventDefault();
  });

  els.stage.addEventListener("keyup", (event) => {
    if (!isSessionActive()) {
      return;
    }
    sendControl({ kind: "key", key: event.key, pressed: false });
    event.preventDefault();
  });
}

function bindMobileControls() {
  els.stage.addEventListener("pointerdown", (event) => {
    if (event.pointerType !== "touch" || !isSessionActive()) {
      return;
    }
    event.preventDefault();

    const touch = {
      pointerId: event.pointerId,
      startX: event.clientX,
      startY: event.clientY,
      x: event.clientX,
      y: event.clientY,
      startAt: Date.now(),
      lastMoveAt: 0,
      maxMove: 0,
      suppressTap: false
    };
    state.mobileTouches.set(event.pointerId, touch);

    try {
      els.stage.setPointerCapture(event.pointerId);
    } catch (_) {
      // noop
    }

    const pos = normalizedPositionFromClient(event.clientX, event.clientY);
    sendControl({ kind: "mouse_move", x: pos.x, y: pos.y });

    if (state.mobileDragMode) {
      handleDragPointerDown(event.pointerId);
      return;
    }

    if (state.mobileTouches.size === 1) {
      state.mobileGesture = "single";
      state.mobilePrimaryPointerId = event.pointerId;
      state.mobileLongPressFired = false;
      startMobileLongPress(event.pointerId);
      return;
    }

    if (state.mobileTouches.size === 2) {
      enterScrollGesture();
    }
  });

  els.stage.addEventListener("pointermove", (event) => {
    if (event.pointerType !== "touch" || !state.mobileTouches.has(event.pointerId) || !isSessionActive()) {
      return;
    }
    event.preventDefault();

    const touch = state.mobileTouches.get(event.pointerId);
    touch.x = event.clientX;
    touch.y = event.clientY;
    const move = Math.hypot(touch.x - touch.startX, touch.y - touch.startY);
    touch.maxMove = Math.max(touch.maxMove, move);
    state.mobileTouches.set(event.pointerId, touch);

    if (state.mobileDragMode) {
      handleDragPointerMove(event.pointerId);
      return;
    }

    if (state.mobileGesture === "scroll" && state.mobileTouches.size >= 2) {
      const center = currentTouchCenter();
      if (state.mobileScrollCenter) {
        const deltaY = center.y - state.mobileScrollCenter.y;
        handleMobileScrollDelta(deltaY);
      }
      state.mobileScrollCenter = center;
      return;
    }

    if (state.mobileGesture === "single" && event.pointerId === state.mobilePrimaryPointerId) {
      if (touch.maxMove > 12) {
        cancelMobileLongPress();
      }
      const now = Date.now();
      if (now - touch.lastMoveAt < 34) {
        return;
      }
      touch.lastMoveAt = now;
      state.mobileTouches.set(event.pointerId, touch);

      const pos = normalizedPositionFromClient(event.clientX, event.clientY);
      sendControl({ kind: "mouse_move", x: pos.x, y: pos.y });
    }
  });

  const finishTouch = (event) => {
    if (event.pointerType !== "touch" || !state.mobileTouches.has(event.pointerId)) {
      return;
    }
    event.preventDefault();

    const touch = state.mobileTouches.get(event.pointerId);

    try {
      els.stage.releasePointerCapture(event.pointerId);
    } catch (_) {
      // noop
    }

    state.mobileTouches.delete(event.pointerId);

    if (state.mobileDragMode) {
      handleDragPointerUp(event.pointerId);
      return;
    }

    if (state.mobileGesture === "scroll") {
      if (state.mobileTouches.size < 2) {
        resetMobileGestureState();
      }
      return;
    }

    if (event.pointerId === state.mobilePrimaryPointerId) {
      cancelMobileLongPress();
      maybeSendSingleTouchClick(touch);
      resetMobileGestureState();
    }
  };

  els.stage.addEventListener("pointerup", finishTouch);
  els.stage.addEventListener("pointercancel", finishTouch);
}

function bindMobileGestureActions() {
  els.mobileDragToggleBtn.addEventListener("click", () => {
    state.mobileDragMode = !state.mobileDragMode;

    if (!state.mobileDragMode) {
      forceReleaseMobileDrag();
      setStatus(t("dragModeDisabled"));
    } else {
      setStatus(t("dragModeEnabled"));
    }

    resetMobileGestureState();
    updateMobileDragToggleButton();
  });
}

function handleDragPointerDown(pointerId) {
  if (state.mobileDragPressed) {
    return;
  }
  if (state.mobileTouches.size !== 1) {
    return;
  }
  state.mobileDragPointerId = pointerId;
  state.mobileDragPressed = true;
  sendControl({ kind: "mouse_click", button: "left", pressed: true });
}

function handleDragPointerMove(pointerId) {
  if (!state.mobileDragPressed || pointerId !== state.mobileDragPointerId) {
    return;
  }
  const touch = state.mobileTouches.get(pointerId);
  if (!touch) {
    return;
  }
  const now = Date.now();
  if (now - touch.lastMoveAt < 24) {
    return;
  }
  touch.lastMoveAt = now;
  state.mobileTouches.set(pointerId, touch);

  const pos = normalizedPositionFromClient(touch.x, touch.y);
  sendControl({ kind: "mouse_move", x: pos.x, y: pos.y });
}

function handleDragPointerUp(pointerId) {
  if (!state.mobileDragPressed || pointerId !== state.mobileDragPointerId) {
    return;
  }
  sendControl({ kind: "mouse_click", button: "left", pressed: false });
  state.mobileDragPressed = false;
  state.mobileDragPointerId = null;
}

function forceReleaseMobileDrag() {
  if (!state.mobileDragPressed) {
    return;
  }
  sendControl({ kind: "mouse_click", button: "left", pressed: false });
  state.mobileDragPressed = false;
  state.mobileDragPointerId = null;
}

function updateMobileDragToggleButton() {
  if (!els.mobileDragToggleBtn) {
    return;
  }
  els.mobileDragToggleBtn.textContent = state.mobileDragMode ? t("dragOn") : t("dragOff");
  els.mobileDragToggleBtn.classList.toggle("mobile-toggle-on", state.mobileDragMode);
}

function enterScrollGesture() {
  cancelMobileLongPress();
  state.mobileGesture = "scroll";
  state.mobilePrimaryPointerId = null;
  state.mobileLongPressFired = false;
  state.mobileScrollCenter = currentTouchCenter();
  state.mobileScrollRemainderY = 0;

  for (const touch of state.mobileTouches.values()) {
    touch.suppressTap = true;
  }
}

function currentTouchCenter() {
  const touches = Array.from(state.mobileTouches.values()).slice(0, 2);
  if (touches.length === 0) {
    return { x: 0, y: 0 };
  }
  const sum = touches.reduce(
    (acc, item) => {
      acc.x += item.x;
      acc.y += item.y;
      return acc;
    },
    { x: 0, y: 0 }
  );
  return {
    x: sum.x / touches.length,
    y: sum.y / touches.length
  };
}

function handleMobileScrollDelta(deltaY) {
  state.mobileScrollRemainderY += deltaY;
  const threshold = 26;
  let steps = 0;

  while (Math.abs(state.mobileScrollRemainderY) >= threshold) {
    if (state.mobileScrollRemainderY > 0) {
      steps += 1;
      state.mobileScrollRemainderY -= threshold;
    } else {
      steps -= 1;
      state.mobileScrollRemainderY += threshold;
    }
  }

  if (steps !== 0) {
    sendControl({ kind: "wheel", dx: 0, dy: -steps });
  }
}

function startMobileLongPress(pointerId) {
  cancelMobileLongPress();

  state.mobileLongPressTimer = window.setTimeout(() => {
    if (!isSessionActive()) {
      return;
    }
    if (state.mobileGesture !== "single") {
      return;
    }
    if (state.mobilePrimaryPointerId !== pointerId) {
      return;
    }
    if (state.mobileTouches.size !== 1) {
      return;
    }

    const touch = state.mobileTouches.get(pointerId);
    if (!touch || touch.maxMove > 14) {
      return;
    }

    const pos = normalizedPositionFromClient(touch.x, touch.y);
    sendControl({ kind: "mouse_move", x: pos.x, y: pos.y });
    sendControl({ kind: "mouse_click", button: "right", pressed: true });
    sendControl({ kind: "mouse_click", button: "right", pressed: false });

    touch.suppressTap = true;
    state.mobileTouches.set(pointerId, touch);
    state.mobileLongPressFired = true;
  }, 560);
}

function cancelMobileLongPress() {
  if (state.mobileLongPressTimer) {
    window.clearTimeout(state.mobileLongPressTimer);
    state.mobileLongPressTimer = null;
  }
}

function maybeSendSingleTouchClick(touch) {
  if (!touch || touch.suppressTap) {
    return;
  }

  const pressDuration = Date.now() - touch.startAt;
  if (touch.maxMove > 18) {
    return;
  }

  if (state.mobileLongPressFired || pressDuration >= 560) {
    sendControl({ kind: "mouse_click", button: "right", pressed: true });
    sendControl({ kind: "mouse_click", button: "right", pressed: false });
    return;
  }

  sendControl({ kind: "mouse_click", button: "left", pressed: true });
  sendControl({ kind: "mouse_click", button: "left", pressed: false });
}

function resetMobileGestureState() {
  cancelMobileLongPress();
  state.mobileGesture = "idle";
  state.mobilePrimaryPointerId = null;
  state.mobileLongPressFired = false;
  state.mobileScrollCenter = null;
  state.mobileScrollRemainderY = 0;
}

function bindMobileKeyboardBridge() {
  els.mobileKeyboardBtn.addEventListener("click", () => {
    if (!state.isMobileMode) {
      return;
    }
    els.mobileImeInput.focus();
    setStatus(t("softKeyboardOpened"));
  });

  els.mobileImeInput.addEventListener("beforeinput", (event) => {
    if (!isSessionActive()) {
      return;
    }

    if (event.inputType === "deleteContentBackward") {
      state.imePendingBackspace += 1;
      return;
    }

    if (event.inputType === "insertParagraph" || event.inputType === "insertLineBreak") {
      sendKeyTap("Enter");
      event.preventDefault();
    }
  });

  els.mobileImeInput.addEventListener("input", () => {
    if (!isSessionActive()) {
      state.imePrevValue = els.mobileImeInput.value;
      return;
    }

    if (state.imePendingBackspace > 0) {
      for (let i = 0; i < state.imePendingBackspace; i += 1) {
        sendKeyTap("Backspace");
      }
      state.imePendingBackspace = 0;
    }

    const nextValue = els.mobileImeInput.value;
    const added = findAddedText(state.imePrevValue, nextValue);
    if (added) {
      sendControl({ kind: "text", text: added });
    }

    state.imePrevValue = nextValue;
    if (nextValue.length > 64) {
      els.mobileImeInput.value = "";
      state.imePrevValue = "";
    }
  });
}

async function loadRuntimeConfig() {
  const res = await fetch("/api/runtime", {
    headers: {
      Authorization: `Bearer ${state.token}`
    }
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || t("runtimeConfigFailed"));
  }
  applyWebRTCConfig(data.webrtc || {});
}

async function connectWS() {
  return new Promise((resolve, reject) => {
    const protocol = location.protocol === "https:" ? "wss" : "ws";
    const ws = new WebSocket(`${protocol}://${location.host}/ws`);

    ws.onopen = () => {
      ws.send(JSON.stringify({ type: "auth", token: state.token }));
      state.ws = ws;
      resolve();
    };

    ws.onmessage = async (event) => {
      let msg;
      try {
        msg = JSON.parse(event.data);
      } catch (_) {
        return;
      }
      await handleWSMessage(msg);
    };

    ws.onerror = () => {
      reject(new Error(t("websocketConnectFailed")));
    };

    ws.onclose = () => {
      setStatus(t("signalDisconnected"));
      stopSession(false);
    };
  });
}

async function handleWSMessage(msg) {
  switch (msg.type) {
    case "ws_ready":
      if (msg.webrtc) {
        applyWebRTCConfig(msg.webrtc);
      }
      break;
    case "error":
      setStatus(`${t("errorPrefix")}: ${msg.message}`);
      break;
    case "devices_updated":
      await loadDevices();
      break;
    case "devices":
      state.devices = msg.devices || [];
      renderDevices();
      break;
    case "session_created":
      if (msg.webrtc) {
        applyWebRTCConfig(msg.webrtc);
      }
      state.sessionId = msg.session_id;
      startP2PWatchdogIfNeeded();
      setStatus(t("sessionCreated"));
      renderTransportStatus();
      break;
    case "session_updated":
      setStatus(`${t("sessionConfigApplied")}: ${msg.quality || "medium"} / ${msg.fps || 15}fps`);
      break;
    case "signal":
      if (!state.sessionId) {
        state.sessionId = msg.session_id;
      }
      await handleSignal(msg.payload || {});
      break;
    case "session_ended":
      setStatus(`${t("sessionEnded")}: ${msg.reason || t("transportUnknown")}`);
      stopSession(false);
      break;
    default:
      break;
  }
}

function applyWebRTCConfig(config) {
  const mode = String(config.mode || "p2p").toLowerCase();
  state.configuredTransportMode = mode === "turn" ? "turn" : "p2p";
  state.configuredForceRelay = Boolean(config.force_relay);

  const servers = Array.isArray(config.ice_servers) ? config.ice_servers : [];
  state.configuredIceServers = servers
    .map((server) => {
      const urls = Array.isArray(server.urls) ? server.urls : typeof server.urls === "string" ? [server.urls] : [];
      if (urls.length === 0) {
        return null;
      }
      return {
        urls,
        username: server.username || undefined,
        credential: server.credential || undefined
      };
    })
    .filter(Boolean);

  if (state.configuredIceServers.length === 0 && state.configuredTransportMode === "p2p") {
    state.configuredIceServers = [{ urls: ["stun:stun.l.google.com:19302"] }];
  }

  renderTransportStatus();
}

async function loadDevices() {
  if (!state.token) {
    return;
  }
  const res = await fetch("/api/devices", {
    headers: {
      Authorization: `Bearer ${state.token}`
    }
  });
  const data = await res.json();
  if (!res.ok) {
    setStatus(`${t("loadDevicesFailed")}: ${data.error || t("transportUnknown")}`);
    return;
  }
  state.devices = data.devices || [];

  if (state.selectedDeviceId && !state.devices.some((device) => device.id === state.selectedDeviceId)) {
    state.selectedDeviceId = "";
  }
  if (!state.selectedDeviceId && state.devices.length > 0) {
    state.selectedDeviceId = state.devices[0].id;
  }

  renderDevices();
  populateDisplayOptions();
}

function renderDevices() {
  const container = els.deviceList;
  container.innerHTML = "";

  if (state.devices.length === 0) {
    const div = document.createElement("div");
    div.className = "device-item";
    div.textContent = t("noOnlineDevices");
    container.appendChild(div);
    return;
  }

  state.devices.forEach((device) => {
    const item = document.createElement("div");
    item.className = "device-item";
    if (device.id === state.selectedDeviceId) {
      item.classList.add("active");
    }

    item.innerHTML = `
      <strong>${escapeHtml(device.name)}</strong>
      <small>ID: ${escapeHtml(device.id)}</small>
      <small>${escapeHtml(device.os || t("transportUnknown"))}</small>
      <small>${(device.displays || []).length} ${t("displaysSuffix")}</small>
    `;

    item.addEventListener("click", () => {
      state.selectedDeviceId = device.id;
      renderDevices();
      populateDisplayOptions();
    });

    container.appendChild(item);
  });
}

function populateDisplayOptions() {
  const device = state.devices.find((item) => item.id === state.selectedDeviceId);
  els.displaySelect.innerHTML = "";

  if (!device || !Array.isArray(device.displays) || device.displays.length === 0) {
    const option = document.createElement("option");
    option.value = "1";
    option.textContent = t("defaultDisplay");
    els.displaySelect.appendChild(option);
    return;
  }

  device.displays.forEach((display) => {
    const option = document.createElement("option");
    option.value = String(display.id);
    option.textContent = `${display.label} (${display.width}x${display.height})`;
    els.displaySelect.appendChild(option);
  });
}

async function startSession() {
  if (!state.selectedDeviceId) {
    setStatus(t("selectDeviceFirst"));
    return;
  }
  if (!state.ws || state.ws.readyState !== WebSocket.OPEN) {
    setStatus(t("signalUnavailable"));
    return;
  }

  if (state.pc) {
    stopSession(false);
  }

  setupPeerConnection();
  state.sessionId = "";
  state.activeTransportMode = "unknown";
  state.connectedOnce = false;
  state.p2pWarningSent = false;

  const payload = {
    type: "start_session",
    device_id: state.selectedDeviceId,
    display_id: Number(els.displaySelect.value || "1"),
    quality: els.qualitySelect.value,
    fps: Number(els.fpsInput.value)
  };

  state.ws.send(JSON.stringify(payload));
  setStatus(t("sessionRequestSent"));
  renderTransportStatus();
}

function setupPeerConnection() {
  clearTransportProbe();

  const rtcConfig = {
    iceServers: state.configuredIceServers
  };
  if (state.configuredForceRelay) {
    rtcConfig.iceTransportPolicy = "relay";
  }

  state.pc = new RTCPeerConnection(rtcConfig);

  state.pc.ontrack = (event) => {
    const [stream] = event.streams;
    if (stream) {
      els.remoteVideo.srcObject = stream;
    } else {
      els.remoteVideo.srcObject = new MediaStream([event.track]);
    }
    els.stage.classList.add("connected");
    setStatus(t("mediaConnected"));
    scheduleTransportProbe();
  };

  state.pc.onicecandidate = (event) => {
    if (!event.candidate || !state.sessionId) {
      return;
    }
    sendSignal({ type: "candidate", candidate: event.candidate });
  };

  state.pc.oniceconnectionstatechange = () => {
    const iceState = state.pc?.iceConnectionState || "unknown";
    if (iceState === "connected" || iceState === "completed") {
      state.connectedOnce = true;
      clearP2PWatchdog();
      scheduleTransportProbe();
    }

    if ((iceState === "failed" || iceState === "disconnected") && state.configuredTransportMode === "p2p") {
      maybeWarnP2PFailure("ice_state_failed", { ice_state: iceState });
    }
  };

  state.pc.onconnectionstatechange = () => {
    const connState = state.pc?.connectionState || "unknown";
    if (connState === "connected") {
      state.connectedOnce = true;
      clearP2PWatchdog();
      scheduleTransportProbe();
      return;
    }

    if (connState === "failed" || connState === "disconnected" || connState === "closed") {
      if (state.configuredTransportMode === "p2p") {
        maybeWarnP2PFailure("connection_state_failed", { connection_state: connState });
      }
      setStatus(`${t("webrtcState")}: ${connState}`);
      stopSession(false);
    }
  };
}

async function handleSignal(payload) {
  if (!state.pc) {
    setupPeerConnection();
  }

  if (payload.type === "offer") {
    await state.pc.setRemoteDescription(payload);
    const answer = await state.pc.createAnswer();
    await state.pc.setLocalDescription(answer);
    sendSignal(state.pc.localDescription);
    return;
  }

  if (payload.type === "answer") {
    await state.pc.setRemoteDescription(payload);
    return;
  }

  if (payload.type === "candidate" && payload.candidate) {
    try {
      await state.pc.addIceCandidate(payload.candidate);
    } catch (_) {
      // Ignore candidate race.
    }
  }
}

function sendSignal(payload) {
  if (!state.ws || state.ws.readyState !== WebSocket.OPEN || !state.sessionId) {
    return;
  }
  state.ws.send(
    JSON.stringify({
      type: "signal",
      session_id: state.sessionId,
      payload
    })
  );
}

function sendControl(event) {
  if (!isSessionActive()) {
    return;
  }
  state.ws.send(
    JSON.stringify({
      type: "control",
      session_id: state.sessionId,
      event
    })
  );
}

function sendSessionUpdate() {
  if (!isSessionActive()) {
    return;
  }
  const quality = els.qualitySelect.value;
  const fps = Number(els.fpsInput.value);

  state.ws.send(
    JSON.stringify({
      type: "update_session",
      session_id: state.sessionId,
      quality,
      fps
    })
  );
  sendControl({ kind: "stream_config", quality, fps });
}

function scheduleSessionConfigUpdate(delayMs = 180) {
  if (state.sessionUpdateTimer) {
    window.clearTimeout(state.sessionUpdateTimer);
    state.sessionUpdateTimer = null;
  }

  state.sessionUpdateTimer = window.setTimeout(() => {
    state.sessionUpdateTimer = null;
    sendSessionUpdate();
  }, delayMs);
}

function sendClientWarning(code, message, details = {}) {
  if (!state.ws || state.ws.readyState !== WebSocket.OPEN) {
    return;
  }
  state.ws.send(
    JSON.stringify({
      type: "client_warning",
      session_id: state.sessionId,
      code,
      message,
      details
    })
  );
}

function startP2PWatchdogIfNeeded() {
  clearP2PWatchdog();
  if (state.configuredTransportMode !== "p2p") {
    return;
  }

  state.p2pWatchdogTimer = window.setTimeout(() => {
    if (!isSessionActive() || state.connectedOnce) {
      return;
    }
    maybeWarnP2PFailure("timeout", {
      timeout_ms: 15000,
      note: "P2P likely blocked by NAT/firewall"
    });
  }, 15000);
}

function clearP2PWatchdog() {
  if (state.p2pWatchdogTimer) {
    window.clearTimeout(state.p2pWatchdogTimer);
    state.p2pWatchdogTimer = null;
  }
}

function maybeWarnP2PFailure(code, details) {
  if (state.configuredTransportMode !== "p2p" || state.p2pWarningSent) {
    return;
  }
  state.p2pWarningSent = true;

  const message = t("p2pFailedHint");
  console.warn("[warning]", message, details);
  setStatus(message);
  sendClientWarning(`p2p_${code}`, message, details);
}

function scheduleTransportProbe() {
  clearTransportProbe();

  const runProbe = async () => {
    const mode = await detectActiveTransportMode();
    if (!mode) {
      return;
    }
    state.activeTransportMode = mode;
    renderTransportStatus();

    if (mode !== "unknown") {
      clearTransportProbe();
    }
  };

  runProbe();
  state.transportProbeTimer = window.setInterval(runProbe, 2000);
}

function clearTransportProbe() {
  if (state.transportProbeTimer) {
    window.clearInterval(state.transportProbeTimer);
    state.transportProbeTimer = null;
  }
}

async function detectActiveTransportMode() {
  if (!state.pc) {
    return null;
  }

  let stats;
  try {
    stats = await state.pc.getStats();
  } catch (_) {
    return null;
  }

  let pair = null;
  for (const report of stats.values()) {
    if (report.type === "transport" && report.selectedCandidatePairId && stats.get(report.selectedCandidatePairId)) {
      pair = stats.get(report.selectedCandidatePairId);
      break;
    }
  }

  if (!pair) {
    for (const report of stats.values()) {
      if (report.type === "candidate-pair" && report.state === "succeeded" && report.nominated) {
        pair = report;
        break;
      }
    }
  }

  if (!pair) {
    return "unknown";
  }

  const local = stats.get(pair.localCandidateId);
  const remote = stats.get(pair.remoteCandidateId);
  const localType = String(local?.candidateType || "");
  const remoteType = String(remote?.candidateType || "");

  if (localType === "relay" || remoteType === "relay") {
    return "turn";
  }
  if (localType || remoteType) {
    return "p2p";
  }
  return "unknown";
}

function renderTransportStatus() {
  const configured = state.configuredTransportMode === "turn" ? "TURN" : "P2P";
  const active =
    state.activeTransportMode === "turn"
      ? t("transportTURNRelay")
      : state.activeTransportMode === "p2p"
        ? t("transportP2PDirect")
        : state.activeTransportMode === "unknown"
          ? t("transportDetecting")
          : t("transportIdle");
  els.transportText.textContent = `${t("transportTitle")}: ${t("transportConfigured")}=${configured}, ${t("transportActive")}=${active}`;
}

function stopSession(notifyRemote) {
  if (notifyRemote && state.ws && state.ws.readyState === WebSocket.OPEN && state.sessionId) {
    state.ws.send(JSON.stringify({ type: "session_end", session_id: state.sessionId }));
  }

  if (state.mobileDragPressed && state.ws && state.ws.readyState === WebSocket.OPEN && state.sessionId) {
    sendControl({ kind: "mouse_click", button: "left", pressed: false });
  }

  if (state.pc) {
    try {
      state.pc.close();
    } catch (_) {
      // noop
    }
  }

  clearP2PWatchdog();
  clearTransportProbe();
  if (state.sessionUpdateTimer) {
    window.clearTimeout(state.sessionUpdateTimer);
    state.sessionUpdateTimer = null;
  }

  state.pc = null;
  state.sessionId = "";
  state.connectedOnce = false;
  state.p2pWarningSent = false;
  state.activeTransportMode = "none";

  state.mobileTouches.clear();
  resetMobileGestureState();
  state.mobileDragPressed = false;
  state.mobileDragPointerId = null;

  state.imePendingBackspace = 0;
  state.imePrevValue = "";
  els.mobileImeInput.value = "";
  els.remoteVideo.srcObject = null;
  els.stage.classList.remove("connected");
  renderTransportStatus();
}

function setStatus(text) {
  els.statusText.textContent = text;
}

function isSessionActive() {
  return Boolean(state.sessionId && state.ws && state.ws.readyState === WebSocket.OPEN);
}

function mapMouseButton(value) {
  if (value === 0) {
    return "left";
  }
  if (value === 1) {
    return "middle";
  }
  if (value === 2) {
    return "right";
  }
  return "left";
}

function normalizedPositionFromClient(clientX, clientY) {
  const videoFrame = getRenderedVideoFrameRect();
  const x = clamp((clientX - videoFrame.left) / Math.max(videoFrame.width, 1), 0, 1);
  const y = clamp((clientY - videoFrame.top) / Math.max(videoFrame.height, 1), 0, 1);
  return { x, y };
}

function getRenderedVideoFrameRect() {
  const rect = els.remoteVideo.getBoundingClientRect();
  const videoWidth = els.remoteVideo.videoWidth;
  const videoHeight = els.remoteVideo.videoHeight;

  if (!videoWidth || !videoHeight || rect.width === 0 || rect.height === 0) {
    return {
      left: rect.left,
      top: rect.top,
      width: rect.width,
      height: rect.height
    };
  }

  const streamAspect = videoWidth / videoHeight;
  const boxAspect = rect.width / rect.height;

  if (boxAspect > streamAspect) {
    const height = rect.height;
    const width = height * streamAspect;
    return {
      left: rect.left + (rect.width - width) / 2,
      top: rect.top,
      width,
      height
    };
  }

  const width = rect.width;
  const height = width / streamAspect;
  return {
    left: rect.left,
    top: rect.top + (rect.height - height) / 2,
    width,
    height
  };
}

function sendKeyTap(key) {
  sendControl({ kind: "key", key, pressed: true });
  sendControl({ kind: "key", key, pressed: false });
}

function findAddedText(prev, next) {
  if (next.length === 0) {
    return "";
  }
  if (next.startsWith(prev)) {
    return next.slice(prev.length);
  }

  let prefix = 0;
  while (prefix < prev.length && prefix < next.length && prev[prefix] === next[prefix]) {
    prefix += 1;
  }

  let prevSuffix = prev.length - 1;
  let nextSuffix = next.length - 1;
  while (prevSuffix >= prefix && nextSuffix >= prefix && prev[prevSuffix] === next[nextSuffix]) {
    prevSuffix -= 1;
    nextSuffix -= 1;
  }

  return next.slice(prefix, nextSuffix + 1);
}

function detectMobileMode() {
  if (window.matchMedia && window.matchMedia("(hover: none) and (pointer: coarse)").matches) {
    return true;
  }
  return /Android|iPhone|iPad|iPod|Mobile/i.test(navigator.userAgent || "");
}

function clamp(v, min, max) {
  return Math.max(min, Math.min(max, v));
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}
