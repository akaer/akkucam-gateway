const i18n = window.i18n;
const setText = (element, key, params) => i18n.setText(element, key, params);
const $ = selector => document.querySelector(selector);

const VIEWS = ['auto', 'focus', 'comfort'];
const VIEW_STORAGE_KEY = 'akkucam-wall-view';
const METRICS_STORAGE_KEY = 'akkucam-wall-metrics';
const CANVAS_PADDING = 12;
const GAP_DESKTOP = 12;
const GAP_MOBILE = 8;
const MIN_TILE_DESKTOP = 220;
const MIN_TILE_MOBILE = 140;
const DEFAULT_COMFORT_CHROME = 96;
const FOCUS_MOBILE_BREAKPOINT = 760;
const FOCUS_STRIP_MOBILE_HEIGHT = 118;
const FOCUS_THUMB_INTERVAL_MS = 1000;

let layoutFrame = 0;
let focusThumbTimer = 0;

const state = {
  lang: document.documentElement.dataset.lang || 'en',
  devices: [],
  // Loaded base stations, keyed by hubKey(). Each entry keeps its own channel
  // list and stream map so several hubs can stream at the same time.
  hubs: new Map(),
  // Security codes live in RAM only, keyed by base station DID (ADR-0025).
  credentials: new Map(),
  codeRequests: new Map(),
  codeTarget: null,
  codeResolver: null,
  manualHub: null,
  notice: { key: 'notice.ready', params: {}, type: 'info' },
  // Optional retry handler for the current notice (rendered as a button).
  noticeRetry: null,
  // The setup page is only shown before the first hub is loaded or when the
  // user explicitly adds another base station (ADR-0027).
  setupMode: false,
  view: readStoredView(),
  filter: 'all',
  activeCamera: null,
  details: null,
  liveMetrics: readStoredLiveMetrics(),
};

const ui = {
  body: document.body,
  topbar: $('.topbar'),
  notice: $('#notice'),
  noticeAction: $('#notice-action'),
  setupSection: $('#setup-section'),
  wallSection: $('#wall-section'),
  setupEmpty: $('#setup-empty'),
  backToWall: $('#back-to-wall'),
  discover: $('#discover'),
  hubsSection: $('#hubs-section'),
  hubs: $('#hubs'),
  diagnostics: $('#diagnostics'),
  diagnosticDevices: $('#diagnostic-devices'),
  manualFallback: $('#manual-fallback'),
  forgetCode: $('#forget-code'),
  wallCanvas: $('#wall-canvas'),
  wallGrid: $('#wall-grid'),
  focusLayout: $('#focus-layout'),
  focusMain: $('#focus-main'),
  focusStrip: $('#focus-strip'),
  hubFilters: $('#hub-filters'),
  mobileHubFilters: $('#mobile-hub-filters'),
  viewButtons: Array.from(document.querySelectorAll('[data-view]')),
  startAllHubs: $('#start-all-hubs'),
  stopAllHubs: $('#stop-all-hubs'),
  manageHubs: $('#manage-hubs'),
  wallSettings: $('#wall-settings'),
  mobileWallBar: $('#mobile-wall-bar'),
  mobileFilterButton: $('#mobile-filter-button'),
  mobileFilterLabel: $('#mobile-filter-label'),
  mobileFilterCount: $('#mobile-filter-count'),
  mobileStreamButton: $('#mobile-stream-button'),
  mobileStreamLabel: $('#mobile-stream-label'),
  mobileStreamIcon: $('#mobile-stream-icon'),
  mobileMenuButton: $('#mobile-menu-button'),
  mobileWallMenu: $('#mobile-wall-menu'),
  mobileSheetStart: $('#mobile-sheet-start'),
  mobileSheetStop: $('#mobile-sheet-stop'),
  mobileManageHubs: $('#mobile-manage-hubs'),
  mobileSettings: $('#mobile-settings'),
  settingsDialog: $('#settings-dialog'),
  settingsRadios: Array.from(document.querySelectorAll('input[name="wall-view"]')),
  settingsNote: $('#settings-storage-note'),
  resetWallSettings: $('#reset-wall-settings'),
  detailsDialog: $('#details-dialog'),
  detailsBody: $('#details-body'),
  hubsDialog: $('#hubs-dialog'),
  hubManager: $('#hub-manager'),
  addHub: $('#add-hub'),
  forgetAllCodes: $('#forget-all-codes'),
  settingsLiveMetrics: $('#settings-live-metrics'),
  themeToggle: $('#theme-toggle'),
  languageSelect: $('#language-select'),
  helpToggle: $('#help-toggle'),
  helpDialog: $('#help-dialog'),
  licensesDialog: $('#licenses-dialog'),
  openLicenses: $('#open-licenses'),
  streamHolder: $('#stream-holder'),
  favicon: $('#favicon'),
  codeDialog: $('#code-dialog'),
  codeDialogHub: $('#code-dialog-hub'),
  securityCode: $('#security-code'),
  codeVisibility: $('#code-visibility'),
  codeSubmit: $('#code-submit'),
};

function readStorage(key) {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

function writeStorage(key, value) {
  try {
    localStorage.setItem(key, value);
  } catch {
    // Storage can be unavailable; the current page still works.
  }
}

function removeStorage(key) {
  try {
    localStorage.removeItem(key);
  } catch {
    // Ignore storage failures.
  }
}

function storedView() {
  const value = readStorage(VIEW_STORAGE_KEY);
  return VIEWS.includes(value) ? value : null;
}

function readStoredView() {
  return storedView() || 'auto';
}

// Default is on; only "off" is stored, so the setting stays sparse (ADR-0031).
function readStoredLiveMetrics() {
  return readStorage(METRICS_STORAGE_KEY) !== 'off';
}

function escapeHtml(value) {
  return String(value ?? '').replace(/[&<>'"]/g, character => ({'&':'&amp;','<':'&lt;','>':'&gt;',"'":'&#39;','"':'&quot;'})[character]);
}

function setRawText(element, text) {
  if (!element) return;
  delete element.dataset.i18n;
  delete element.dataset.i18nParams;
  element.textContent = text;
}

// hubKey identifies a base station for credential storage. The DID is stable
// across discovery runs; the IP is only a fallback when the DID is missing.
function hubKey(hub) {
  return hub?.did || hub?.ip || '';
}

function hubLabel(hub) {
  if (!hub) return '';
  return [hub.model, hub.did].filter(Boolean).join(' · ') || i18n.t('hub.fallbackName');
}

// The full DID is too long for a filter chip or a tile badge. The last DID
// segment stays unique enough to distinguish several hubs in normal setups.
function hubShortLabel(hub) {
  if (!hub) return i18n.t('hub.fallbackName');
  const did = String(hub.did || '');
  const suffix = did ? did.split('-').filter(Boolean).pop() : hub.ip;
  return [hub.model, suffix].filter(Boolean).join(' · ') || i18n.t('hub.fallbackName');
}

function cameraId(key, channel) {
  return `${key}::${channel}`;
}

function channelDisplayName(entry, index) {
  const channel = entry?.channels.find(item => item.index === index);
  return channel?.name || i18n.t('camera.fallbackName', { index: index + 1 });
}

function findEntry(key) {
  return state.hubs.get(key) || null;
}

function findChannel(entry, index) {
  return entry?.channels.find(channel => channel.index === index) || null;
}

function allCameras() {
  const entries = [];
  for (const entry of state.hubs.values()) {
    for (const channel of entry.channels) {
      entries.push({ entry, channel, id: cameraId(entry.key, channel.index) });
    }
  }
  return entries;
}

function visibleCameras() {
  const entries = allCameras();
  return state.filter === 'all'
    ? entries
    : entries.filter(({ entry }) => entry.key === state.filter);
}

function streamFor(entry, channel) {
  return entry?.streams.get(channel.index) || null;
}

function cameraState(entry, channel) {
  return streamFor(entry, channel)?.state || 'ready';
}

function streamCount() {
  let count = 0;
  for (const entry of state.hubs.values()) count += entry.streams.size;
  return count;
}

function channelCount() {
  let count = 0;
  for (const entry of state.hubs.values()) count += entry.channels.length;
  return count;
}

function sameCamera(a, b) {
  return Boolean(a && b && a.key === b.key && a.index === b.index);
}

function tileElement(key, index) {
  return [...document.querySelectorAll('.camera-tile')]
    .find(tile => tile.dataset.hubKey === key && Number(tile.dataset.channel) === index) || null;
}

function focusItemElement(key, index) {
  return [...document.querySelectorAll('.focus-item')]
    .find(item => item.dataset.hubKey === key && Number(item.dataset.channel) === index) || null;
}

function setNotice(key, params = {}, type = 'info', retry = null) {
  state.notice = { key, params, type };
  state.noticeRetry = retry;
  renderNotice();
}

function setNoticeError(key, error, extra = {}, retry = null) {
  state.notice = {
    key,
    params: { ...extra, detail: errorText(error) },
    type: 'error',
    errorCode: error && error.code,
  };
  state.noticeRetry = retry;
  renderNotice();
}

function setNoticeRaw(message, type = 'info') {
  state.notice = { raw: message, type };
  state.noticeRetry = null;
  renderNotice();
}

function renderNotice() {
  const text = ui.notice.querySelector('.notice-text');
  if (!text) return;
  const notice = state.notice;
  ui.notice.classList.toggle('error', notice.type === 'error');
  if (ui.noticeAction) ui.noticeAction.hidden = typeof state.noticeRetry !== 'function';
  if (notice.raw !== undefined) {
    setRawText(text, notice.raw);
    ui.notice.title = notice.raw;
    return;
  }
  const params = { ...notice.params };
  if (notice.errorCode && i18n.has(`error.code.${notice.errorCode}`)) {
    params.detail = i18n.t(`error.code.${notice.errorCode}`);
  }
  i18n.setText(text, notice.key, params);
  ui.notice.title = text.textContent;
}

function errorText(error) {
  if (!error) return '';
  const code = error.code;
  if (code && i18n.has(`error.code.${code}`)) return i18n.t(`error.code.${code}`);
  return error.message || String(error);
}

function updateStreamNotice() {
  const active = [];
  for (const entry of state.hubs.values()) {
    for (const channel of entry.streams.keys()) {
      active.push({ entry, channel });
    }
  }
  const count = active.length;
  if (!count) return;
  // The whole sentence lives in the catalogs; the channel list is formatted
  // here because it contains untranslated device names (ADR-0021, ADR-0028).
  const names = active.map(({ entry, channel }) => channelDisplayName(entry, channel));
  const channels = new Intl.ListFormat(state.lang, { style: 'short', type: 'conjunction' }).format(names);
  setNotice('notice.streamsActive', { count, channels });
}

async function api(path, options = {}) {
  const response = await fetch(path, options);
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    const error = new Error(body.error || `HTTP ${response.status}`);
    error.code = body.code;
    error.status = response.status;
    throw error;
  }
  return body;
}

function storedTheme() {
  const value = readStorage('akkucam-theme');
  return value === 'light' || value === 'dark' ? value : null;
}

function systemTheme() {
  return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function renderedTheme() {
  return document.documentElement.dataset.theme || systemTheme();
}

// theme is "light", "dark" or null to follow the system preference (ADR-0013).
function applyTheme(theme) {
  if (theme) document.documentElement.dataset.theme = theme;
  else delete document.documentElement.dataset.theme;
  const meta = document.querySelector('meta[name="color-scheme"]');
  if (meta) meta.content = theme || 'light dark';
  const rendered = theme || systemTheme();
  const next = rendered === 'dark' ? 'light' : 'dark';
  const key = next === 'dark' ? 'theme.toDark' : 'theme.toLight';
  ui.themeToggle.dataset.i18nTitle = key;
  ui.themeToggle.title = i18n.t(key);
  if (ui.favicon) {
    ui.favicon.href = rendered === 'dark' ? '/icon-dark.svg' : '/icon-light.svg';
  }
}

function setTheme(theme) {
  if (theme) writeStorage('akkucam-theme', theme);
  else removeStorage('akkucam-theme');
  applyTheme(theme);
}

// Native closedby="any" handles light dismiss where supported. Safari lacks it,
// so the fallback tests the pointer position instead of event.target, which
// would also match the dialog's own padding (ADR-0030).
function wireDialogs() {
  if ('closedBy' in HTMLDialogElement.prototype) return;
  const outside = (event, dialog) => {
    const rect = dialog.getBoundingClientRect();
    return event.clientX < rect.left || event.clientX > rect.right
      || event.clientY < rect.top || event.clientY > rect.bottom;
  };
  for (const dialog of document.querySelectorAll('dialog[closedby="any"]')) {
    let pressedOutside = false;
    dialog.addEventListener('pointerdown', event => { pressedOutside = outside(event, dialog); });
    dialog.addEventListener('pointerup', event => {
      if (pressedOutside && outside(event, dialog)) dialog.close();
    });
  }
}

/* -------------------------------------------------------------------------
 * Layout: setup vs live wall
 * ---------------------------------------------------------------------- */

function isSetupVisible() {
  return state.setupMode || state.hubs.size === 0;
}

function renderLayout() {
  const setup = isSetupVisible();
  if (setup) stopFocusThumbTimer();
  renderViewButtons();
  document.body.classList.toggle('wall-active', !setup);
  ui.setupSection.hidden = !setup;
  ui.wallSection.hidden = setup;
  ui.backToWall.hidden = state.hubs.size === 0;
  if (setup) {
    renderHubs();
    if (state.devices.length) renderDiagnostics();
    renderNotice();
  } else {
    renderWall();
  }
}

function showSetupForAdd() {
  if (streamCount()) {
    setNotice('notice.stopBeforeAddHub', {}, 'error');
    return;
  }
  ui.hubsDialog.close();
  state.setupMode = true;
  renderLayout();
  window.scrollTo({ top: 0, behavior: 'instant' });
  discover();
}

function backToWall() {
  state.setupMode = false;
  renderLayout();
}

/* -------------------------------------------------------------------------
 * Discovery, hub setup and credentials
 * ---------------------------------------------------------------------- */

function renderHubs() {
  const hubs = state.devices.filter(device => device.model === 'SMH-T1R' || String(device.did).startsWith('CMAG-'));
  ui.hubsSection.hidden = hubs.length === 0;
  ui.setupEmpty.hidden = hubs.length > 0;
  ui.hubs.innerHTML = hubs.map(device => {
    const key = hubKey(device);
    const loaded = state.hubs.has(key);
    const metadata = loaded ? renderHubMetadata(key) : '';
    const subtitleParts = [device.did, device.ip].filter(Boolean);
    const hasCode = state.credentials.has(key);
    const streams = state.hubs.get(key)?.streams.size || 0;
    return `
      <article class="hub-card">
        <div class="hub-icon" aria-hidden="true"><svg class="icon" focusable="false"><use href="#icon-camera"></use></svg></div>
        <div><h3>${escapeHtml(device.model || i18n.t('hub.fallbackName'))}</h3>${subtitleParts.length ? `<p>${escapeHtml(subtitleParts.join(' · '))}</p>` : ''}</div>
        <div class="hub-actions">
          <button class="primary load-channels" data-hub-key="${escapeHtml(key)}" data-ip="${escapeHtml(device.ip)}" data-did="${escapeHtml(device.did)}" ${streams ? 'disabled' : ''} data-i18n="${loaded ? 'hub.reloadChannels' : 'hub.loadChannels'}">${escapeHtml(i18n.t(loaded ? 'hub.reloadChannels' : 'hub.loadChannels'))}</button>
          <span class="hub-credential" data-hub-key="${escapeHtml(key)}" ${hasCode ? '' : 'hidden'}>
            <span class="credential-state active" data-i18n="hub.codeSaved">Code saved</span>
            <button class="text-button forget-hub-code" type="button" data-hub-key="${escapeHtml(key)}" data-i18n="hub.forgetCode">Remove code</button>
          </span>
        </div>
        ${metadata}
      </article>`;
  }).join('');
  i18n.applyTranslations(ui.hubs);
  updateGlobalControls();
  if (hubs.length === 0 && state.devices.length) setNotice('notice.noSupportedHub', {}, 'error');
}

function renderHubMetadata(key) {
  const entry = state.hubs.get(key);
  if (!entry) return '';
  const info = entry.info || {};
  const sdTotal = info.sdTotalBytes;
  const sdFree = info.sdFreeBytes;
  const sdUsed = (typeof sdTotal === 'number' && typeof sdFree === 'number' && sdTotal >= sdFree) ? (sdTotal - sdFree) : null;
  const sdStatus = info.sdCardStatus;
  // DID, IP and model already stand in the card header; repeating them here
  // would bury the values that are only available in this grid (ADR-0015).
  const items = [
    metadataItem('hub.meta.firmware', 'Firmware', info.firmware),
    metadataItem('hub.meta.mcu', 'MCU', info.mcuVersion),
    metadataItem('hub.meta.mac', 'MAC address', info.mac || entry.hub?.mac),
    sdStatus === 'noCard' ? metadataItemKey('hub.meta.sdCard', 'SD card', 'hub.meta.sdNone', 'No card') : '',
    sdTotal ? metadataItem('hub.meta.sdTotal', 'SD total', formatDecimalBytes(sdTotal)) : '',
    sdFree !== undefined && sdFree !== null && sdTotal ? metadataItem('hub.meta.sdFree', 'SD free', formatDecimalBytes(sdFree)) : '',
    sdUsed !== null && sdTotal ? metadataItem('hub.meta.sdUsed', 'SD used', formatDecimalBytes(sdUsed)) : '',
  ].filter(Boolean);
  if (!items.length) return '';
  return `<div class="hub-metadata">${items.join('')}</div>`;
}

function metadataItem(labelKey, label, value) {
  if (value === undefined || value === null || value === '') return '';
  return `<div><span data-i18n="${labelKey}">${escapeHtml(label)}</span><strong>${escapeHtml(value)}</strong></div>`;
}

function metadataItemKey(labelKey, label, valueKey, value) {
  return `<div><span data-i18n="${labelKey}">${escapeHtml(label)}</span><strong data-i18n="${valueKey}">${escapeHtml(value)}</strong></div>`;
}

function detailMetadataItem(labelKey, label, value) {
  if (value === undefined || value === null || value === '') return '';
  return `<div><dt data-i18n="${labelKey}">${escapeHtml(label)}</dt><dd>${escapeHtml(value)}</dd></div>`;
}

function formatDecimalBytes(bytes) {
  const value = Number(bytes);
  if (!Number.isFinite(value)) return '';
  const units = [['TB', 1e12], ['GB', 1e9], ['MB', 1e6], ['kB', 1e3]];
  for (const [unit, threshold] of units) {
    if (value >= threshold) {
      const digits = value >= threshold * 10 ? 0 : 1;
      return `${new Intl.NumberFormat(state.lang, { maximumFractionDigits: digits }).format(value / threshold)} ${unit}`;
    }
  }
  return `${new Intl.NumberFormat(state.lang).format(value)} B`;
}

// Durations arrive as machine values from the API (ADR-0021); the dashboard
// is the only place that formats them for the active language.
function formatMilliseconds(milliseconds) {
  const value = Number(milliseconds);
  if (!Number.isFinite(value) || value < 0) return '';
  const seconds = value / 1000;
  const formatted = new Intl.NumberFormat(state.lang, {
    maximumFractionDigits: seconds < 10 ? 1 : 0,
  }).format(seconds);
  return `${formatted}\u00a0s`;
}

function formatDuration(seconds) {
  const total = Math.floor(Number(seconds));
  if (!Number.isFinite(total) || total < 0) return '';
  const hours = Math.floor(total / 3600);
  const minutes = Math.floor((total % 3600) / 60);
  const rest = total % 60;
  const parts = [];
  if (hours) parts.push(i18n.t('duration.hours', { value: hours }));
  if (minutes) parts.push(i18n.t('duration.minutes', { value: minutes }));
  if (!hours && (rest || !parts.length)) parts.push(i18n.t('duration.seconds', { value: rest }));
  return parts.join(' ');
}

function renderDiagnostics() {
  if (!state.devices.length) {
    ui.diagnosticDevices.innerHTML = '<p class="muted" data-i18n="diagnostics.empty">No devices found.</p>';
    i18n.applyTranslations(ui.diagnosticDevices);
    return;
  }
  ui.diagnosticDevices.innerHTML = state.devices.map(device => {
    const values = [device.ip && `${device.ip}${device.ppcsPort ? `:${device.ppcsPort}` : ''}`, device.model, device.mac].filter(Boolean);
    return `<div class="diagnostic-row">${device.did ? `<strong>${escapeHtml(device.did)}</strong>` : ''}${values.map(value => `<span>${escapeHtml(value)}</span>`).join('')}</div>`;
  }).join('');
}

async function discover() {
  if (streamCount()) {
    setNotice('notice.stopBeforeDiscovery', {}, 'error');
    return;
  }
  ui.discover.disabled = true;
  setNotice('notice.discovering');
  try {
    const result = await api('/api/discover', {method: 'POST'});
    state.devices = result.devices || [];
    renderDiagnostics();
    renderHubs();
    setNotice('notice.devicesFound', { count: state.devices.length, duration: formatMilliseconds(result.durationMs) });
  } catch (error) {
    setNoticeError('notice.discoveryFailed', error, {}, discover);
  } finally {
    ui.discover.disabled = streamCount() > 0;
  }
}

async function loadChannels(hub, button) {
  const key = hubKey(hub);
  const existing = state.hubs.get(key);
  if (streamCount()) {
    setNotice('notice.stopBeforeReload', {}, 'error');
    return;
  }
  const code = await requestSecurityCode(hub);
  if (!code) return;
  button.disabled = true;
  setText(button, 'hub.loadingChannels');
  setNotice('notice.loadingChannels');
  try {
    const result = await api('/api/channels', {
      method: 'POST', headers: {'Content-Type':'application/json'},
      body: JSON.stringify({ip: hub.ip, did: hub.did, securityCode: code}),
    });
    const channels = (result.channels || []).map(channel => ({ ...channel, quality: channel.quality || '1' }));
    if (!channels.length) throw new Error(i18n.t('notice.noChannels'));
    state.hubs.set(key, {
      key,
      hub,
      info: result.hub || {},
      channels,
      streams: existing ? existing.streams : new Map(),
    });
    state.filter = 'all';
    state.setupMode = false;
    ui.manualFallback.hidden = true;
    renderLayout();
    renderHubs();
    updateCredentialBadges();
    setNotice('notice.channelsLoaded', { count: channels.length, hub: hubShortLabel(hub) });
  } catch (error) {
    if (error.code === 'unauthorized') forgetCredential(key, { silent: true });
    state.manualHub = hub;
    ui.manualFallback.hidden = false;
    ui.diagnostics.open = true;
    setNoticeError('notice.channelsFailed', error);
  } finally {
    if (button.isConnected) {
      button.disabled = streamCount() > 0;
      setText(button, state.hubs.has(key) ? 'hub.reloadChannels' : 'hub.loadChannels');
    }
    updateGlobalControls();
  }
}

// requestSecurityCode returns the code for one base station. It prompts at most
// once per DID at a time, so "Start all" does not stack several dialogs.
function requestSecurityCode(hub) {
  const key = hubKey(hub);
  if (state.credentials.has(key)) return Promise.resolve(state.credentials.get(key));
  if (state.codeRequests.has(key)) return state.codeRequests.get(key);
  const request = promptForCode(hub).finally(() => state.codeRequests.delete(key));
  state.codeRequests.set(key, request);
  return request;
}

// A modal dialog can only be opened once. Global actions may request codes for
// several base stations at the same time, so prompts are queued one after the
// other instead of stacking (ADR-0026).
let promptChain = Promise.resolve();

function promptForCode(hub) {
  const request = promptChain.then(() => openCodeDialog(hub));
  promptChain = request.catch(() => {});
  return request;
}

function openCodeDialog(hub) {
  const { promise, resolve } = Promise.withResolvers();
  state.codeTarget = hub;
  state.codeResolver = resolve;
  setText(ui.codeDialogHub, 'dialog.hub', { hub: hubLabel(hub) });
  ui.codeDialog.returnValue = '';
  ui.securityCode.value = '';
  setCodeVisibility(false);
  ui.codeDialog.showModal();
  requestAnimationFrame(() => {
    ui.securityCode.focus();
    ui.securityCode.select();
  });
  return promise;
}

function setCodeVisibility(visible) {
  ui.securityCode.type = visible ? 'text' : 'password';
  ui.codeVisibility.setAttribute('aria-pressed', String(visible));
}

// Firefox compatibility: Enter in an <input> inside <form method="dialog"> may close the dialog
// without setting a meaningful returnValue. We force the expected action here.
ui.codeDialog.querySelector('form')?.addEventListener('submit', event => {
  event.preventDefault();
  // When Enter triggers submit, submitter can be null. Treat it as OK.
  const submitter = event.submitter;
  if (submitter && submitter.value === 'cancel') {
    ui.codeDialog.close('cancel');
    return;
  }
  if (!ui.securityCode.reportValidity()) return;
  ui.codeDialog.close('ok');
});

ui.codeDialog.addEventListener('close', () => {
  const target = state.codeTarget;
  const key = hubKey(target);
  const code = ui.codeDialog.returnValue === 'ok' ? ui.securityCode.value : null;
  if (code && key) {
    state.credentials.set(key, code);
    refreshCredentialViews();
  }
  const resolve = state.codeResolver;
  state.codeTarget = null;
  state.codeResolver = null;
  ui.securityCode.value = '';
  ui.codeDialog.returnValue = '';
  setCodeVisibility(false);
  resolve?.(code);
});

function updateCredentialBadges() {
  for (const badge of ui.hubs.querySelectorAll('.hub-credential')) {
    badge.hidden = !state.credentials.has(badge.dataset.hubKey);
  }
}

function refreshCredentialViews() {
  updateCredentialBadges();
  if (ui.hubsDialog.open) renderHubManager();
}

function forgetCredential(key, options = {}) {
  if (!key) return;
  state.credentials.delete(key);
  refreshCredentialViews();
  if (!options.silent) setNotice('notice.forgotCode', { hub: key });
}

/* -------------------------------------------------------------------------
 * Wall rendering
 * ---------------------------------------------------------------------- */

function setView(view, persist = true) {
  if (!VIEWS.includes(view)) return;
  state.view = view;
  if (persist) writeStorage(VIEW_STORAGE_KEY, view);
  renderViewButtons();
  if (!ui.wallSection.hidden) renderWall();
  renderSettings();
}

function setFilter(key, options = {}) {
  if (key !== 'all' && !state.hubs.has(key)) key = 'all';
  state.filter = key;
  if (state.details && key !== 'all' && state.details.key !== key) {
    closeDetails();
  }
  if (!ui.wallSection.hidden) renderWall();
  if (options.closeSheet) ui.mobileWallMenu.close();
}

function renderFilters() {
  if (state.filter !== 'all' && !state.hubs.has(state.filter)) {
    state.filter = 'all';
  }
  const total = allCameras().length;
  const chips = [
    `<button type="button" class="filter-chip" data-filter="all" aria-pressed="${state.filter === 'all'}">${escapeHtml(i18n.t('wall.allHubs'))} <span>${total}</span></button>`,
    ...[...state.hubs.values()].map(entry => (
      `<button type="button" class="filter-chip" data-filter="${escapeHtml(entry.key)}" aria-pressed="${state.filter === entry.key}" title="${escapeHtml(hubLabel(entry.hub))}">${escapeHtml(hubShortLabel(entry.hub))} <span>${entry.channels.length}</span></button>`
    )),
  ];
  const markup = chips.join('');
  ui.hubFilters.innerHTML = markup;
  ui.mobileHubFilters.innerHTML = markup;
  updateMobileFilter();
}

function updateMobileFilter() {
  const active = state.filter === 'all'
    ? null
    : state.hubs.get(state.filter) || null;
  const count = active ? active.channels.length : allCameras().length;
  if (active) {
    setRawText(ui.mobileFilterLabel, hubShortLabel(active.hub));
  } else {
    setText(ui.mobileFilterLabel, 'wall.allHubs');
  }
  ui.mobileFilterCount.textContent = count;
  const hub = active ? hubShortLabel(active.hub) : i18n.t('wall.allHubs');
  ui.mobileFilterButton.setAttribute('aria-label', i18n.t('wall.mobileFilter', { hub, count }));
}

function renderViewButtons() {
  for (const button of ui.viewButtons) {
    button.setAttribute('aria-pressed', String(button.dataset.view === state.view));
  }
  for (const radio of ui.settingsRadios) {
    radio.checked = radio.value === state.view;
  }
  document.body.dataset.view = state.view;
}

function renderWall() {
  if (ui.wallSection.hidden) return;
  syncMetricsVisibility();
  // Move video elements to the offscreen holder before replacing the wall.
  // The media elements and their HLS instances survive the re-render and keep
  // decoding so the focus thumbnails can keep drawing frames.
  for (const entry of state.hubs.values()) {
    for (const stream of entry.streams.values()) {
      if (!stream.video) continue;
      stream.video.hidden = false;
      ui.streamHolder.append(stream.video);
    }
  }

  renderFilters();
  renderViewButtons();
  updateGlobalControls();

  const entries = visibleCameras();
  if (state.view === 'focus') {
    renderFocus(entries);
  } else {
    renderGrid(entries);
  }
  attachVisibleVideos(entries);

  if (state.details) {
    renderDetailsDrawer();
  }
}

function renderGrid(entries) {
  stopFocusThumbTimer();
  ui.focusLayout.hidden = true;
  ui.focusMain.replaceChildren();
  ui.focusStrip.replaceChildren();
  ui.wallGrid.hidden = false;
  ui.wallGrid.replaceChildren();
  if (!entries.length) {
    ui.wallGrid.innerHTML = `<div class="wall-empty" data-i18n="wall.noCameras">${escapeHtml(i18n.t('wall.noCameras'))}</div>`;
    ui.wallGrid.style.gridTemplateColumns = '1fr';
    return;
  }
  ui.wallGrid.innerHTML = entries.map(({ entry, channel }) => tileMarkup(entry, channel, state.view === 'comfort' ? 'comfort' : 'overlay')).join('');
  i18n.applyTranslations(ui.wallGrid);
  applyGridLayout(entries);
}

function renderFocus(entries) {
  ui.wallGrid.hidden = true;
  ui.wallGrid.replaceChildren();
  ui.focusLayout.hidden = false;
  if (!entries.length) {
    stopFocusThumbTimer();
    ui.focusMain.innerHTML = `<div class="wall-empty" data-i18n="wall.noCameras">${escapeHtml(i18n.t('wall.noCameras'))}</div>`;
    ui.focusStrip.replaceChildren();
    return;
  }

  const active = entries.find(({ entry, channel }) => sameCamera(state.activeCamera, { key: entry.key, index: channel.index })) || entries[0];
  state.activeCamera = { key: active.entry.key, index: active.channel.index };
  ui.focusMain.innerHTML = tileMarkup(active.entry, active.channel, 'overlay');
  ui.focusStrip.innerHTML = entries.map(({ entry, channel }) => focusItemMarkup(entry, channel)).join('');
  i18n.applyTranslations(ui.focusMain);
  i18n.applyTranslations(ui.focusStrip);
  layoutFocus(entries.length);
  startFocusThumbTimer();
}

function paintFocusThumbs() {
  if (ui.wallSection.hidden || state.view !== 'focus') {
    stopFocusThumbTimer();
    return;
  }
  for (const item of ui.focusStrip.querySelectorAll('.focus-item')) {
    const entry = findEntry(item.dataset.hubKey);
    const channel = findChannel(entry, Number(item.dataset.channel));
    const stream = streamFor(entry, channel);
    const canvas = item.querySelector('.focus-thumb-canvas');
    const placeholder = item.querySelector('.focus-thumb-placeholder');
    if (!canvas || !placeholder) continue;
    const live = stream?.state === 'live' && stream.video;
    if (!live) {
      canvas.hidden = true;
      placeholder.hidden = false;
      continue;
    }
    canvas.hidden = false;
    placeholder.hidden = true;
    const video = stream.video;
    if (video.readyState < 2 || !video.videoWidth || !video.videoHeight) continue;
    const context = canvas.getContext('2d');
    if (!context) continue;
    try {
      const scale = Math.max(canvas.width / video.videoWidth, canvas.height / video.videoHeight);
      const sourceWidth = canvas.width / scale;
      const sourceHeight = canvas.height / scale;
      context.drawImage(
        video,
        (video.videoWidth - sourceWidth) / 2,
        (video.videoHeight - sourceHeight) / 2,
        sourceWidth,
        sourceHeight,
        0,
        0,
        canvas.width,
        canvas.height,
      );
    } catch {
      // A frame can be unavailable while the video is still waking up.
    }
  }
}

function startFocusThumbTimer() {
  paintFocusThumbs();
  if (focusThumbTimer) return;
  focusThumbTimer = window.setInterval(paintFocusThumbs, FOCUS_THUMB_INTERVAL_MS);
}

function stopFocusThumbTimer() {
  if (!focusThumbTimer) return;
  window.clearInterval(focusThumbTimer);
  focusThumbTimer = 0;
}

function statePillMarkup(entry, channel) {
  const stateName = cameraState(entry, channel);
  let key = 'camera.state.ready';
  let className = '';
  if (stateName === 'live') {
    key = 'camera.state.live';
    className = 'live';
  } else if (stateName === 'starting') {
    key = 'camera.state.connecting';
    className = 'connecting';
  } else if (stateName === 'waking') {
    key = 'camera.state.waking';
    className = 'connecting';
  }
  return `<span class="state-pill ${className}"><i></i><span data-i18n="${key}">${escapeHtml(i18n.t(key))}</span></span>`;
}

function cameraNameMarkup(channel) {
  if (channel.name) {
    return `<span class="tile-name">${escapeHtml(channel.name)}</span>`;
  }
  const fallback = i18n.t('camera.fallbackName', { index: channel.index + 1 });
  return `<span class="tile-name" data-i18n="camera.fallbackName" data-i18n-params='{"index":${channel.index + 1}}'>${escapeHtml(fallback)}</span>`;
}

function qualitySelect(entry, channel) {
  const id = escapeHtml(cameraId(entry.key, channel.index));
  const options = [
    ['1', 'camera.quality.fullHd', 'Full HD'],
    ['4', 'camera.quality.auto', 'Auto'],
    ['2', 'camera.quality.hd', 'HD'],
    ['3', 'camera.quality.sd', 'SD'],
  ].map(([value, key, label]) => `<option value="${value}" ${channel.quality === value ? 'selected' : ''} data-i18n="${key}">${escapeHtml(i18n.t(key))}</option>`).join('');
  return `<select class="quality" data-camera-id="${id}">${options}</select>`;
}

function liveMetricValues(stream) {
  const status = stream?.status;
  if (!status) return [];
  const values = [];
  if (Number.isFinite(status.bitrateBytesPerSecond)) {
    const kbps = new Intl.NumberFormat(state.lang, { maximumFractionDigits: 0 }).format((status.bitrateBytesPerSecond * 8) / 1000);
    values.push(i18n.t('camera.metric.kbpsValue', { value: kbps }));
  }
  if (status.framesPerSecond) {
    values.push(i18n.t('camera.metric.fpsValue', { value: status.framesPerSecond }));
  }
  if (status.battery?.capacity !== undefined) {
    values.push(i18n.t('camera.metric.percentValue', { value: status.battery.capacity }));
  }
  return values;
}

function tileMetricsMarkup(entry, channel) {
  const text = liveMetricValues(streamFor(entry, channel)).join(' · ');
  return `<span class="tile-metrics" aria-hidden="true">${escapeHtml(text)}</span>`;
}

function fullscreenButtonMarkup(entry, channel, extraClass = '') {
  if (cameraState(entry, channel) !== 'live') return '';
  return `<button class="tile-action icon ${extraClass}" type="button" data-action="fullscreen" aria-label="${escapeHtml(i18n.t('camera.fullscreen'))}" data-i18n-aria-label="camera.fullscreen"><svg class="icon" aria-hidden="true" focusable="false"><use href="#icon-expand"></use></svg></button>`;
}

function placeholderMarkup(entry, channel) {
  const stateName = cameraState(entry, channel);
  if (stateName === 'live') {
    return `<div class="video-placeholder" hidden><strong data-i18n="camera.placeholder.title">${escapeHtml(i18n.t('camera.placeholder.title'))}</strong><span data-i18n="camera.placeholder.hint">${escapeHtml(i18n.t('camera.placeholder.hint'))}</span></div>`;
  }
  if (stateName === 'starting' || stateName === 'waking') {
    return `<div class="video-placeholder"><strong data-i18n="camera.state.connecting">${escapeHtml(i18n.t('camera.state.connecting'))}</strong><span data-i18n="camera.placeholder.connectingHint">${escapeHtml(i18n.t('camera.placeholder.connectingHint'))}</span></div>`;
  }
  return `<div class="video-placeholder"><strong data-i18n="camera.placeholder.title">${escapeHtml(i18n.t('camera.placeholder.title'))}</strong><span data-i18n="camera.placeholder.hint">${escapeHtml(i18n.t('camera.placeholder.hint'))}</span></div>`;
}

function motionControlMarkup(entry, channel) {
  if (channel.motionEnabled !== true && channel.motionEnabled !== false) return '';
  const disabled = streamCount() > 0;
  return `<label class="switch-row compact"><span data-i18n="camera.motion">${escapeHtml(i18n.t('camera.motion'))}</span><span class="switch"><input class="motion-toggle" type="checkbox" data-camera-id="${escapeHtml(cameraId(entry.key, channel.index))}" ${channel.motionEnabled ? 'checked' : ''} ${disabled ? 'disabled' : ''} title="${disabled ? escapeHtml(i18n.t('notice.lockedByStreams')) : ''}"><i aria-hidden="true"></i></span></label>`;
}

function streamButtonMarkup(entry, channel) {
  const stateName = cameraState(entry, channel);
  if (stateName === 'starting' || stateName === 'waking') {
    return `<button class="primary small stream-toggle" type="button" data-action="stream" data-camera-id="${escapeHtml(cameraId(entry.key, channel.index))}" disabled data-i18n="camera.wakingButton">${escapeHtml(i18n.t('camera.wakingButton'))}</button>`;
  }
  const live = stateName === 'live';
  return `<button class="${live ? 'secondary' : 'primary'} small stream-toggle" type="button" data-action="stream" data-camera-id="${escapeHtml(cameraId(entry.key, channel.index))}" data-i18n="${live ? 'action.stop' : 'action.start'}">${escapeHtml(i18n.t(live ? 'action.stop' : 'action.start'))}</button>`;
}

function tileMarkup(entry, channel, style) {
  const id = cameraId(entry.key, channel.index);
  const stateName = cameraState(entry, channel);
  const live = stateName === 'live';
  const label = `${channelDisplayName(entry, channel.index)}, ${hubLabel(entry.hub)}`;
  const index = `<span class="tile-index">${channel.index + 1}</span>`;

  if (style === 'comfort') {
    return `<article class="tile camera-tile is-comfort" data-hub-key="${escapeHtml(entry.key)}" data-channel="${channel.index}" data-camera-id="${escapeHtml(id)}" tabindex="0" role="group" aria-label="${escapeHtml(label)}">
      <header class="tile-head">
        ${index}
        <span class="tile-titles"><strong>${escapeHtml(channelDisplayName(entry, channel.index))}</strong><span>${escapeHtml(hubShortLabel(entry.hub))}</span></span>
        ${statePillMarkup(entry, channel)}
      </header>
      <div class="tile-media">
        ${placeholderMarkup(entry, channel)}
        ${tileMetricsMarkup(entry, channel)}
        ${fullscreenButtonMarkup(entry, channel, 'media-action')}
      </div>
      <footer class="tile-foot">
        <label class="compact-field"><span data-i18n="camera.quality">${escapeHtml(i18n.t('camera.quality'))}</span>${qualitySelect(entry, channel)}</label>
        ${motionControlMarkup(entry, channel)}
        ${streamButtonMarkup(entry, channel)}
      </footer>
    </article>`;
  }

  return `<article class="tile camera-tile" data-hub-key="${escapeHtml(entry.key)}" data-channel="${channel.index}" data-camera-id="${escapeHtml(id)}" tabindex="0" role="group" aria-label="${escapeHtml(label)}">
    <div class="tile-media">
      ${placeholderMarkup(entry, channel)}
      ${tileMetricsMarkup(entry, channel)}
      <div class="tile-overlay">
        <div class="tile-overlay-top">
          <div class="tile-ident">${index}${cameraNameMarkup(channel)}</div>
          ${statePillMarkup(entry, channel)}
        </div>
        <div class="tile-overlay-bottom">
          ${fullscreenButtonMarkup(entry, channel)}
        </div>
      </div>
    </div>
  </article>`;
}

function focusItemMarkup(entry, channel) {
  const id = cameraId(entry.key, channel.index);
  const active = sameCamera(state.activeCamera, { key: entry.key, index: channel.index });
  return `<button class="focus-item ${active ? 'is-active' : ''}" type="button" data-hub-key="${escapeHtml(entry.key)}" data-channel="${channel.index}" data-camera-id="${escapeHtml(id)}" aria-pressed="${active}">
    <span class="focus-thumb">
      <canvas class="focus-thumb-canvas" width="320" height="180" hidden aria-hidden="true"></canvas>
      <span class="focus-thumb-placeholder" aria-hidden="true"><svg class="icon icon-lg" focusable="false"><use href="#icon-camera"></use></svg></span>
      ${statePillMarkup(entry, channel)}
    </span>
    <span class="focus-label"><strong>${channel.index + 1} · ${escapeHtml(channelDisplayName(entry, channel.index))}</strong></span>
  </button>`;
}

function measureComfortChrome() {
  const head = ui.wallGrid.querySelector('.tile.is-comfort .tile-head');
  const foot = ui.wallGrid.querySelector('.tile.is-comfort .tile-foot');
  if (!head || !foot) return DEFAULT_COMFORT_CHROME;
  return Math.round(head.getBoundingClientRect().height + foot.getBoundingClientRect().height);
}

function applyGridLayout(entries) {
  const width = ui.wallCanvas.clientWidth - CANVAS_PADDING * 2;
  const height = ui.wallCanvas.clientHeight - CANVAS_PADDING * 2;
  if (width <= 0 || height <= 0 || !entries.length) return;

  const chrome = state.view === 'comfort' ? measureComfortChrome() : 0;
  const layout = computeGridLayout(width, height, entries.length, state.view, chrome);
  ui.wallGrid.style.setProperty('--grid-gap', `${layout.gap}px`);
  ui.wallGrid.style.setProperty('--tile-h', `${layout.tileH}px`);
  ui.wallGrid.style.gridTemplateColumns = `repeat(${layout.cols}, ${layout.tileW}px)`;
  syncFillers(layout.cols * layout.rows - entries.length);

}

function syncFillers(needed) {
  const fillers = Array.from(ui.wallGrid.querySelectorAll('.tile-empty'));
  while (fillers.length < needed) {
    const filler = document.createElement('div');
    filler.className = 'tile tile-empty';
    filler.setAttribute('aria-hidden', 'true');
    filler.dataset.emptyLabel = i18n.t('wall.emptyCell');
    ui.wallGrid.append(filler);
    fillers.push(filler);
  }
  while (fillers.length > needed) {
    fillers.pop().remove();
  }
}

function computeGridLayout(width, height, count, view, chrome) {
  const gap = width < 700 ? GAP_MOBILE : GAP_DESKTOP;
  const minTile = width < 700 ? MIN_TILE_MOBILE : MIN_TILE_DESKTOP;
  const phonePortrait = width < 640 && height > width;

  if (phonePortrait) {
    const cols = 1;
    const rows = Math.ceil(count / cols);
    const tileW = width;
    const tileH = (tileW * 9) / 16 + chrome;
    const totalHeight = rows * tileH + gap * (rows - 1);
    return { cols, rows, tileW, tileH, gap };
  }

  let best = null;
  for (let cols = 1; cols <= count; cols += 1) {
    const rows = Math.ceil(count / cols);
    const widthLimited = (width - gap * (cols - 1)) / cols;
    const heightLimited = (((height - gap * (rows - 1)) / rows) - chrome) * 16 / 9;
    const tileW = Math.min(widthLimited, heightLimited);
    if (tileW <= 0) continue;
    const better = !best
      || tileW > best.tileW + 0.5
      || (Math.abs(tileW - best.tileW) <= 0.5 && rows < best.rows);
    if (better) best = { cols, rows, tileW };
  }

  if (best && best.tileW >= minTile) {
    return {
      cols: best.cols,
      rows: best.rows,
      tileW: best.tileW,
      tileH: (best.tileW * 9) / 16 + chrome,
      gap,
    };
  }

  const cols = Math.max(1, Math.min(count, Math.floor((width + gap) / (minTile + gap))));
  const rows = Math.ceil(count / cols);
  const tileW = (width - gap * (cols - 1)) / cols;
  const tileH = (tileW * 9) / 16 + chrome;
  return {
    cols,
    rows,
    tileW,
    tileH,
    gap,
  };
}

function layoutFocus(count) {
  const width = ui.wallCanvas.clientWidth - CANVAS_PADDING * 2;
  const height = ui.wallCanvas.clientHeight - CANVAS_PADDING * 2;
  const mobile = width < FOCUS_MOBILE_BREAKPOINT;

  if (mobile) {
    const gap = GAP_MOBILE;
    const mainHeight = Math.max(120, height - FOCUS_STRIP_MOBILE_HEIGHT - gap);
    const tileW = Math.min(width, (mainHeight * 16) / 9);
    const tileH = (tileW * 9) / 16;
    ui.focusLayout.style.setProperty('--focus-w', `${tileW}px`);
    ui.focusLayout.style.setProperty('--focus-h', `${tileH}px`);
    ui.focusLayout.style.setProperty('--strip-h', `${FOCUS_STRIP_MOBILE_HEIGHT}px`);
    ui.focusLayout.classList.add('is-mobile');
  } else {
    const gap = GAP_DESKTOP;
    const stripW = Math.max(150, Math.min(230, width * 0.17));
    const mainWidth = width - stripW - gap;
    const tileW = Math.min(mainWidth, (height * 16) / 9);
    const tileH = (tileW * 9) / 16;
    ui.focusLayout.style.setProperty('--focus-w', `${tileW}px`);
    ui.focusLayout.style.setProperty('--focus-h', `${tileH}px`);
    ui.focusLayout.style.setProperty('--strip-w', `${stripW}px`);
    ui.focusLayout.classList.remove('is-mobile');
  }

}

function attachVisibleVideos(entries) {
  for (const { entry, channel } of entries) {
    const stream = streamFor(entry, channel);
    if (!stream || stream.state !== 'live' || !stream.video) continue;
    const tile = tileElement(entry.key, channel.index);
    if (!tile) continue;
    const media = tile.querySelector('.tile-media');
    const placeholder = media?.querySelector('.video-placeholder');
    if (!media) continue;
    if (placeholder) placeholder.hidden = true;
    stream.video.hidden = false;
    stream.video.controls = document.fullscreenElement === stream.video;
    media.append(stream.video);
    stream.video.play().catch(() => {});
  }
}

function scheduleLayout() {
  if (layoutFrame) return;
  layoutFrame = requestAnimationFrame(() => {
    layoutFrame = 0;
    if (ui.wallSection.hidden) return;
    const entries = visibleCameras();
    if (!entries.length) return;
    if (state.view === 'focus') {
      layoutFocus(entries.length);
    } else {
      applyGridLayout(entries);
    }
  });
}

function setActiveCamera(key, index) {
  state.activeCamera = { key, index };
  if (!ui.wallSection.hidden) renderWall();
}

/* -------------------------------------------------------------------------
 * Settings and hub manager
 * ---------------------------------------------------------------------- */

function renderSettings() {
  for (const radio of ui.settingsRadios) {
    radio.checked = radio.value === state.view;
  }
  if (ui.settingsLiveMetrics) ui.settingsLiveMetrics.checked = state.liveMetrics;
  const saved = storedView();
  if (saved) {
    setText(ui.settingsNote, 'settings.saved', { view: i18n.t(`view.${saved}`) });
  } else {
    setText(ui.settingsNote, 'settings.default', { view: i18n.t('view.auto') });
  }
}

function openSettings() {
  renderSettings();
  if (!ui.settingsDialog.open) ui.settingsDialog.showModal();
}

function resetWallSettings() {
  removeStorage(VIEW_STORAGE_KEY);
  setLiveMetrics(true);
  state.view = 'auto';
  state.filter = 'all';
  renderViewButtons();
  if (!ui.wallSection.hidden) renderWall();
  renderSettings();
  setNotice('notice.wallReset');
}

function renderHubManager() {
  if (!state.hubs.size) {
    ui.hubManager.innerHTML = '<p class="muted" data-i18n="setup.emptyTitle">No base station loaded</p>';
    i18n.applyTranslations(ui.hubManager);
    return;
  }
  const rows = [...state.hubs.values()].map(entry => {
    const count = entry.channels.length;
    const live = entry.streams.size;
    const hasCode = state.credentials.has(entry.key);
    const mac = entry.info?.mac || entry.hub?.mac;
    const network = [
      entry.hub?.ip ? `${i18n.t('hub.meta.ip')} ${entry.hub.ip}` : '',
      mac ? `${i18n.t('hub.meta.mac')} ${mac}` : '',
    ].filter(Boolean);
    return `<div class="hub-row" data-hub-key="${escapeHtml(entry.key)}">
      <div><strong>${escapeHtml(hubLabel(entry.hub))}</strong><small>${escapeHtml(i18n.t('workspace.cameraCount', { count }))} · ${escapeHtml(i18n.t('wall.liveCount', { count: live }))}</small>${network.map(part => `<small class="hub-row-net">${escapeHtml(part)}</small>`).join('')}</div>
      <span class="code-state ${hasCode ? 'active' : ''}">${escapeHtml(i18n.t(hasCode ? 'notice.credentialActive' : 'notice.credentialInactive'))}</span>
      <div class="hub-row-actions">
        <button type="button" class="secondary small" data-hub-action="start" data-hub-key="${escapeHtml(entry.key)}" ${live >= count ? 'disabled' : ''} data-i18n="workspace.startAll">${escapeHtml(i18n.t('workspace.startAll'))}</button>
        <button type="button" class="secondary small" data-hub-action="stop" data-hub-key="${escapeHtml(entry.key)}" ${live === 0 ? 'disabled' : ''} data-i18n="workspace.stopAll">${escapeHtml(i18n.t('workspace.stopAll'))}</button>
        <button type="button" class="text-button" data-hub-action="forget" data-hub-key="${escapeHtml(entry.key)}" data-i18n="hub.forgetCode">${escapeHtml(i18n.t('hub.forgetCode'))}</button>
      </div>
    </div>`;
  });
  ui.hubManager.innerHTML = rows.join('');
  i18n.applyTranslations(ui.hubManager);
  ui.addHub.disabled = streamCount() > 0;
  ui.addHub.title = ui.addHub.disabled ? i18n.t('notice.stopBeforeAddHub') : '';
}

function openHubManager() {
  renderHubManager();
  if (!ui.hubsDialog.open) ui.hubsDialog.showModal();
}

/* -------------------------------------------------------------------------
 * Transport controls
 * ---------------------------------------------------------------------- */

function updateGlobalControls() {
  const streams = streamCount();
  const channels = channelCount();
  ui.discover.disabled = streams > 0;
  ui.discover.title = streams > 0 ? i18n.t('notice.lockedByStreams') : '';
  for (const button of document.querySelectorAll('.load-channels')) {
    button.disabled = streams > 0;
  }
  for (const input of document.querySelectorAll('.motion-toggle')) {
    input.disabled = streams > 0;
  }
  for (const button of document.querySelectorAll('[data-manual-channel]')) {
    button.disabled = streams > 0;
  }
  ui.startAllHubs.disabled = channels === 0 || streams >= channels;
  ui.stopAllHubs.disabled = streams === 0;

  const anyLive = streams > 0;
  ui.mobileStreamButton.disabled = channels === 0;
  ui.mobileStreamButton.dataset.action = anyLive ? 'stop' : 'start';
  setText(ui.mobileStreamLabel, anyLive ? 'action.stop' : 'action.start');
  ui.mobileStreamIcon.querySelector('use')?.setAttribute('href', anyLive ? '#icon-stop' : '#icon-play');
  ui.mobileStreamButton.dataset.i18nAriaLabel = anyLive ? 'wall.stopAllStreams' : 'wall.startAllStreams';
  ui.mobileStreamButton.setAttribute('aria-label', i18n.t(anyLive ? 'wall.stopAllStreams' : 'wall.startAllStreams'));

  ui.mobileSheetStart.disabled = channels === 0 || streams >= channels;
  ui.mobileSheetStop.disabled = streams === 0;
  ui.addHub.disabled = streams > 0;

  if (ui.hubsDialog.open) {
    renderHubManager();
  }
}

async function startAll(key) {
  const entry = state.hubs.get(key);
  if (!entry) return;
  const pending = entry.channels.filter(channel => !entry.streams.has(channel.index));
  await Promise.all(pending.map(channel => startChannel(key, channel.index)));
  if (ui.hubsDialog.open) renderHubManager();
  updateGlobalControls();
}

async function startAllHubs() {
  ui.startAllHubs.disabled = true;
  try {
    await Promise.all([...state.hubs.keys()].map(key => startAll(key)));
  } finally {
    updateGlobalControls();
  }
}

async function stopAll(key, notify = true) {
  const entry = state.hubs.get(key);
  if (!entry) return;
  await Promise.all([...entry.streams.keys()].map(channel => stopChannel(key, channel, false)));
  if (!notify) return;
  if (streamCount()) updateStreamNotice();
  else setNotice('notice.allStreamsStopped');
}

async function stopAllHubs() {
  ui.stopAllHubs.disabled = true;
  try {
    await Promise.all([...state.hubs.keys()].map(key => stopAll(key, false)));
    setNotice('notice.allStreamsStopped');
  } finally {
    updateGlobalControls();
  }
}

function ensureVideo(entry, channel) {
  const stream = entry.streams.get(channel.index);
  if (!stream) return null;
  if (!stream.video) {
    const video = document.createElement('video');
    video.controls = false;
    video.muted = true;
    video.playsInline = true;
    video.className = 'stream-video';
    stream.video = video;
    ui.streamHolder.append(video);
  }
  return stream.video;
}

async function startChannel(key, channelIndex) {
  const entry = state.hubs.get(key);
  const channel = findChannel(entry, channelIndex);
  if (!entry || !channel || entry.streams.has(channelIndex)) return;
  const code = await requestSecurityCode(entry.hub);
  if (!code) return;

  const stream = { id: null, hls: null, timer: null, state: 'starting', status: null, video: null };
  entry.streams.set(channelIndex, stream);
  if (!ui.wallSection.hidden) renderWall();
  if (state.details && sameCamera(state.details, { key, index: channelIndex })) renderDetailsDrawer();
  updateGlobalControls();
  setNotice('notice.streamStarting', { channel: channelDisplayName(entry, channelIndex) });

  try {
    const quality = Number(channel.quality || '1');
    const result = await api('/api/live/start', {
      method:'POST', headers:{'Content-Type':'application/json'},
      body:JSON.stringify({ip:entry.hub.ip, did:entry.hub.did, securityCode:code, cameraIndex:channelIndex, quality}),
    });
    stream.id = result.id;
    stream.state = 'waking';
    if (!ui.wallSection.hidden) renderWall();
    if (state.details && sameCamera(state.details, { key, index: channelIndex })) renderDetailsDrawer();

    const playlist = await waitForPlaylist(result.playlistUrl, 25000);
    const video = ensureVideo(entry, channel);
    attachPlayer(video, playlist, stream);
    stream.state = 'live';
    if (!ui.wallSection.hidden) renderWall();
    if (state.details && sameCamera(state.details, { key, index: channelIndex })) renderDetailsDrawer();
    stream.timer = setInterval(() => updateStreamStatus(key, channelIndex), 1000);
    updateStreamStatus(key, channelIndex);
    updateStreamNotice();
    updateGlobalControls();
  } catch (error) {
    if (error.code === 'unauthorized') forgetCredential(key, { silent: true });
    if (stream.id) {
      await api('/api/live/stop', {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({id:stream.id})}).catch(() => {});
    }
    stream.hls?.destroy();
    if (stream.timer) clearInterval(stream.timer);
    if (stream.video) {
      stream.video.pause();
      stream.video.removeAttribute('src');
      stream.video.load();
      stream.video.remove();
    }
    entry.streams.delete(channelIndex);
    if (!ui.wallSection.hidden) renderWall();
    if (state.details && sameCamera(state.details, { key, index: channelIndex })) renderDetailsDrawer();
    setNoticeError('notice.channelError', error, { channel: channelDisplayName(entry, channelIndex) }, () => startChannel(key, channelIndex));
    updateGlobalControls();
  }
}

async function stopChannel(key, channelIndex, notify = true) {
  const entry = state.hubs.get(key);
  const stream = entry?.streams.get(channelIndex);
  if (stream) {
    if (stream.timer) clearInterval(stream.timer);
    stream.hls?.destroy();
    if (stream.video) {
      stream.video.pause();
      stream.video.removeAttribute('src');
      stream.video.load();
      stream.video.remove();
    }
    entry.streams.delete(channelIndex);
    if (stream.id) {
      await api('/api/live/stop', {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({id:stream.id})}).catch(() => {});
    }
  }
  if (!ui.wallSection.hidden) renderWall();
  if (state.details && sameCamera(state.details, { key, index: channelIndex })) renderDetailsDrawer();
  updateGlobalControls();
  if (notify) {
    if (streamCount()) updateStreamNotice();
    else setNotice('notice.noStreams');
  }
}

async function waitForPlaylist(url, timeout) {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    const response = await fetch(`${url}?probe=${Date.now()}`, {cache:'no-store'}).catch(() => null);
    if (response?.ok && (await response.text()).includes('#EXTM3U')) return url;
    await new Promise(resolve => setTimeout(resolve, 500));
  }
  throw new Error(i18n.t('camera.playlistTimeout'));
}

function attachPlayer(video, url, stream) {
  if (!video) return;
  if (window.Hls && Hls.isSupported()) {
    stream.hls = new Hls({liveSyncDurationCount:2, liveMaxLatencyDurationCount:5});
    stream.hls.loadSource(`${url}?t=${Date.now()}`);
    stream.hls.attachMedia(video);
    stream.hls.on(Hls.Events.MANIFEST_PARSED, () => video.play().catch(() => {}));
  } else {
    video.src = `${url}?t=${Date.now()}`;
    video.play().catch(() => {});
  }
}

async function updateStreamStatus(key, channelIndex) {
  const entry = state.hubs.get(key);
  const stream = entry?.streams.get(channelIndex);
  if (!stream) return;
  try {
    const status = await api(`/api/live/status?id=${encodeURIComponent(stream.id)}`);
    if (!status.running) return stopChannel(key, channelIndex, false);
    stream.status = status;
    if (state.details && sameCamera(state.details, { key, index: channelIndex })) {
      updateDetailsMetrics();
    }
    updateTileMetrics(key, channelIndex);
  } catch (_) {}
}

/* -------------------------------------------------------------------------
 * Details drawer
 * ---------------------------------------------------------------------- */

function openDetails(key, channelIndex) {
  const entry = findEntry(key);
  if (!entry || !findChannel(entry, channelIndex)) return;
  state.details = { key, index: channelIndex };
  renderDetailsDrawer();
  if (!ui.detailsDialog.open) ui.detailsDialog.showModal();
}

function closeDetails() {
  if (ui.detailsDialog.open) ui.detailsDialog.close();
  state.details = null;
}

function renderDetailsDrawer() {
  if (!state.details) return;
  const entry = findEntry(state.details.key);
  const channel = findChannel(entry, state.details.index);
  if (!entry || !channel) {
    closeDetails();
    return;
  }
  const stream = streamFor(entry, channel);
  const stateName = cameraState(entry, channel);
  const live = stateName === 'live';
  const connecting = stateName === 'starting' || stateName === 'waking';
  const metadata = [
    detailMetadataItem('camera.meta.channelIndex', 'Channel index', channel.index),
    detailMetadataItem('camera.meta.name', 'Device name', channel.name),
    detailMetadataItem('camera.meta.model', 'Model', channel.model),
    detailMetadataItem('camera.meta.firmware', 'Firmware', channel.fwVersion),
    detailMetadataItem('camera.meta.mcu', 'MCU', channel.mcuVersion),
  ].filter(Boolean).join('');

  const action = live || connecting
    ? `<button type="button" class="secondary" data-detail-action="toggle-stream" ${connecting ? 'disabled' : ''} data-i18n="${connecting ? 'camera.wakingButton' : 'action.stop'}">${escapeHtml(i18n.t(connecting ? 'camera.wakingButton' : 'action.stop'))}</button>`
    : `<button type="button" class="primary" data-detail-action="toggle-stream" data-i18n="action.start">${escapeHtml(i18n.t('action.start'))}</button>`;
  const fullscreen = live
    ? `<button type="button" class="secondary" data-detail-action="fullscreen" data-i18n="camera.fullscreen">${escapeHtml(i18n.t('camera.fullscreen'))}</button>`
    : '';
  const motion = channel.motionEnabled === true || channel.motionEnabled === false
    ? `<label class="switch-row"><span data-i18n="camera.motion">${escapeHtml(i18n.t('camera.motion'))}</span><span class="switch"><input class="motion-toggle" type="checkbox" data-camera-id="${escapeHtml(cameraId(entry.key, channel.index))}" ${channel.motionEnabled ? 'checked' : ''} ${streamCount() > 0 ? 'disabled' : ''} title="${streamCount() > 0 ? escapeHtml(i18n.t('notice.lockedByStreams')) : ''}"><i aria-hidden="true"></i></span></label>`
    : '';

  ui.detailsBody.innerHTML = `
    <div class="dialog-head">
      <div>
        <p class="eyebrow">${escapeHtml(i18n.t('camera.fallbackName', { index: channel.index + 1 }))}</p>
        <h2 id="details-title">${escapeHtml(channelDisplayName(entry, channel.index))}</h2>
        <p class="detail-sub">${escapeHtml(hubLabel(entry.hub))}${entry.hub?.ip ? ` · ${escapeHtml(entry.hub.ip)}` : ''}</p>
      </div>
      <button type="button" class="icon-button" data-close aria-label="${escapeHtml(i18n.t('dialog.close'))}" data-i18n-aria-label="dialog.close"><svg class="icon" aria-hidden="true" focusable="false"><use href="#icon-close"></use></svg></button>
    </div>
    <div class="detail-status">${statePillMarkup(entry, channel)}</div>
    <div class="detail-actions">${action}${fullscreen}</div>
    <div class="detail-fields">
      <label class="field"><span data-i18n="camera.quality">${escapeHtml(i18n.t('camera.quality'))}</span>${qualitySelect(entry, channel)}</label>
      ${motion}
    </div>
    ${metadata ? `<h3 data-i18n="camera.details">${escapeHtml(i18n.t('camera.details'))}</h3><dl class="detail-grid">${metadata}</dl>` : ''}
    <div data-details-metrics>${detailMetricsMarkup(stream)}</div>
    <p class="detail-note" data-i18n="camera.detailsNote">${escapeHtml(i18n.t('camera.detailsNote'))}</p>
  `;
  i18n.applyTranslations(ui.detailsBody);
}

function metricRows(stream) {
  if (!stream?.status) return [];
  const status = stream.status;
  const rows = [];
  if (Number.isFinite(status.bitrateBytesPerSecond)) {
    const kbps = new Intl.NumberFormat(state.lang, { maximumFractionDigits: 0 }).format((status.bitrateBytesPerSecond * 8) / 1000);
    rows.push(['camera.metric.bitrate', i18n.t('camera.metric.kbpsValue', { value: kbps })]);
  }
  if (status.framesPerSecond) {
    rows.push(['camera.metric.fps', i18n.t('camera.metric.fpsValue', { value: status.framesPerSecond })]);
  }
  if (status.battery?.capacity !== undefined) {
    rows.push(['camera.metric.battery', status.battery.charging
      ? i18n.t('camera.battery.charging', { percent: status.battery.capacity })
      : i18n.t('camera.metric.percentValue', { value: status.battery.capacity })]);
  }
  if (Number.isFinite(status.uptimeSeconds) && status.uptimeSeconds > 0) {
    rows.push(['camera.metric.uptime', formatDuration(status.uptimeSeconds)]);
  }
  return rows;
}

function metricRowsMarkup(rows, listClass) {
  return `<dl class="${listClass}">${rows.map(([key, value]) => `<div><dt data-i18n="${key}">${escapeHtml(i18n.t(key))}</dt><dd>${escapeHtml(value)}</dd></div>`).join('')}</dl>`;
}

function detailMetricsMarkup(stream) {
  const rows = metricRows(stream);
  if (!rows.length) return '';
  return `<h3 data-i18n="camera.liveData">${escapeHtml(i18n.t('camera.liveData'))}</h3>${metricRowsMarkup(rows, 'detail-grid')}`;
}

function updateDetailsMetrics() {
  const slot = ui.detailsBody.querySelector('[data-details-metrics]');
  if (!slot || !state.details) return;
  const entry = findEntry(state.details.key);
  const channel = findChannel(entry, state.details.index);
  if (!entry || !channel) return;
  slot.innerHTML = detailMetricsMarkup(streamFor(entry, channel));
  i18n.applyTranslations(slot);
}

function setQuality(key, channelIndex, value) {
  const entry = findEntry(key);
  const channel = findChannel(entry, channelIndex);
  if (!channel) return;
  channel.quality = value;
  for (const select of document.querySelectorAll('select.quality')) {
    if (select.dataset.cameraId === cameraId(key, channelIndex)) select.value = value;
  }
}

async function toggleMotion(key, channelIndex, input) {
  const entry = findEntry(key);
  const channel = findChannel(entry, channelIndex);
  if (!entry || !channel) return;
  const code = await requestSecurityCode(entry.hub);
  if (!code) return;
  const requested = input.checked;
  input.disabled = true;
  setNotice('camera.motion.changing', { channel: channelDisplayName(entry, channelIndex) });
  try {
    const result = await api('/api/motion', {
      method:'POST', headers:{'Content-Type':'application/json'},
      body:JSON.stringify({ip:entry.hub.ip, did:entry.hub.did, securityCode:code, channel:channelIndex, enabled:requested}),
    });
    channel.motionEnabled = result.motionEnabled;
    channel.pirMode = result.pirMode;
    if (!ui.wallSection.hidden) renderWall();
    if (state.details && sameCamera(state.details, { key, index: channelIndex })) renderDetailsDrawer();
    setNotice('notice.motionUpdated', { channel: channelDisplayName(entry, channelIndex), state: i18n.t(channel.motionEnabled ? 'camera.motion.on' : 'camera.motion.off') });
  } catch (error) {
    if (error.code === 'unauthorized') forgetCredential(key, { silent: true });
    input.checked = !requested;
    setNoticeError('notice.motionFailed', error);
  } finally {
    input.disabled = streamCount() > 0;
  }
}

function openFullscreen(key, channelIndex) {
  const stream = state.hubs.get(key)?.streams.get(channelIndex);
  const video = stream?.video || tileElement(key, channelIndex)?.querySelector('video');
  if (!video) return;
  if (ui.detailsDialog.open) ui.detailsDialog.close();
  video.controls = true;
  if (video.requestFullscreen) {
    video.requestFullscreen().catch(() => {});
  } else if (video.webkitEnterFullscreen) {
    video.webkitEnterFullscreen();
  }
}

/* -------------------------------------------------------------------------
 * Live metrics overlay
 * ---------------------------------------------------------------------- */

function updateTileMetrics(key, channelIndex) {
  if (!state.liveMetrics) return;
  const entry = findEntry(key);
  const channel = findChannel(entry, channelIndex);
  if (!entry || !channel) return;
  const text = liveMetricValues(streamFor(entry, channel)).join(' · ');
  for (const tile of document.querySelectorAll('.camera-tile')) {
    if (tile.dataset.hubKey !== key || Number(tile.dataset.channel) !== channelIndex) continue;
    const slot = tile.querySelector('.tile-metrics');
    if (slot && slot.textContent !== text) slot.textContent = text;
  }
}

function syncMetricsVisibility() {
  document.body.classList.toggle('metrics-off', !state.liveMetrics);
}

function setLiveMetrics(enabled) {
  state.liveMetrics = enabled;
  if (enabled) removeStorage(METRICS_STORAGE_KEY);
  else writeStorage(METRICS_STORAGE_KEY, 'off');
  if (ui.settingsLiveMetrics) ui.settingsLiveMetrics.checked = enabled;
  syncMetricsVisibility();
  if (!enabled) return;
  for (const entry of state.hubs.values()) {
    for (const channel of entry.streams.keys()) updateTileMetrics(entry.key, channel);
  }
}

/* -------------------------------------------------------------------------
 * Event wiring
 * ---------------------------------------------------------------------- */

ui.themeToggle.addEventListener('click', () => {
  const target = renderedTheme() === 'dark' ? 'light' : 'dark';
  setTheme(target === systemTheme() ? null : target);
});

ui.codeVisibility.addEventListener('click', () => {
  setCodeVisibility(ui.securityCode.type === 'password');
});

window.matchMedia?.('(prefers-color-scheme: dark)').addEventListener?.('change', () => {
  if (!storedTheme()) applyTheme(null);
});

window.addEventListener('storage', event => {
  if (event.key === 'akkucam-theme') applyTheme(storedTheme());
});
ui.languageSelect.addEventListener('change', () => i18n.setLanguage(ui.languageSelect.value));
ui.helpToggle.addEventListener('click', () => {
  if (!ui.helpDialog.open) ui.helpDialog.showModal();
});
ui.openLicenses.addEventListener('click', () => {
  if (ui.helpDialog.open) ui.helpDialog.close();
  if (!ui.licensesDialog.open) ui.licensesDialog.showModal();
});

ui.discover.addEventListener('click', discover);
ui.backToWall.addEventListener('click', backToWall);
ui.hubs.addEventListener('click', event => {
  const forget = event.target.closest('.forget-hub-code');
  if (forget) {
    forgetCredential(forget.dataset.hubKey);
    return;
  }
  const button = event.target.closest('.load-channels');
  if (!button) return;
  const hub = state.devices.find(device => device.ip === button.dataset.ip && device.did === button.dataset.did);
  if (hub) loadChannels(hub, button);
});

ui.startAllHubs.addEventListener('click', startAllHubs);
ui.stopAllHubs.addEventListener('click', stopAllHubs);

ui.viewButtons.forEach(button => {
  button.addEventListener('click', () => {
    setView(button.dataset.view);
    if (button.closest('#mobile-wall-menu')) ui.mobileWallMenu.close();
  });
});

ui.hubFilters.addEventListener('click', event => {
  const chip = event.target.closest('[data-filter]');
  if (chip) setFilter(chip.dataset.filter);
});

ui.mobileHubFilters.addEventListener('click', event => {
  const chip = event.target.closest('[data-filter]');
  if (chip) setFilter(chip.dataset.filter, { closeSheet: true });
});

ui.mobileFilterButton.addEventListener('click', openMobileMenu);
ui.mobileMenuButton.addEventListener('click', openMobileMenu);
ui.mobileStreamButton.addEventListener('click', () => {
  if (ui.mobileStreamButton.dataset.action === 'stop') stopAllHubs();
  else startAllHubs();
});
ui.mobileSheetStart.addEventListener('click', () => {
  ui.mobileWallMenu.close();
  startAllHubs();
});
ui.mobileSheetStop.addEventListener('click', () => {
  ui.mobileWallMenu.close();
  stopAllHubs();
});
ui.mobileManageHubs.addEventListener('click', () => {
  ui.mobileWallMenu.close();
  openHubManager();
});
ui.mobileSettings.addEventListener('click', () => {
  ui.mobileWallMenu.close();
  openSettings();
});

ui.manageHubs.addEventListener('click', openHubManager);
ui.wallSettings.addEventListener('click', openSettings);
ui.settingsRadios.forEach(radio => {
  radio.addEventListener('change', () => {
    if (radio.checked) setView(radio.value);
  });
});
ui.resetWallSettings.addEventListener('click', resetWallSettings);
ui.settingsLiveMetrics?.addEventListener('change', () => setLiveMetrics(ui.settingsLiveMetrics.checked));
ui.addHub.addEventListener('click', showSetupForAdd);

ui.hubManager.addEventListener('click', event => {
  const button = event.target.closest('[data-hub-action]');
  if (!button) return;
  const key = button.dataset.hubKey;
  if (button.dataset.hubAction === 'start') startAll(key);
  else if (button.dataset.hubAction === 'stop') stopAll(key);
  else if (button.dataset.hubAction === 'forget') forgetCredential(key);
});

ui.detailsBody.addEventListener('click', event => {
  const button = event.target.closest('[data-detail-action]');
  if (!button || !state.details) return;
  if (button.dataset.detailAction === 'toggle-stream') {
    const entry = findEntry(state.details.key);
    const stream = entry?.streams.get(state.details.index);
    if (stream) stopChannel(state.details.key, state.details.index);
    else startChannel(state.details.key, state.details.index);
  }
  if (button.dataset.detailAction === 'fullscreen') {
    openFullscreen(state.details.key, state.details.index);
  }
});

ui.detailsBody.addEventListener('change', event => {
  if (!state.details) return;
  if (event.target.matches('.quality')) {
    setQuality(state.details.key, state.details.index, event.target.value);
  }
  if (event.target.matches('.motion-toggle')) {
    toggleMotion(state.details.key, state.details.index, event.target);
  }
});

ui.detailsDialog.addEventListener('close', () => {
  state.details = null;
});

ui.wallSection.addEventListener('click', event => {
  const focused = event.target.closest('.focus-item');
  if (focused) {
    setActiveCamera(focused.dataset.hubKey, Number(focused.dataset.channel));
    return;
  }
  const tile = event.target.closest('.camera-tile');
  if (!tile) return;
  const key = tile.dataset.hubKey;
  const channel = Number(tile.dataset.channel);
  const actionButton = event.target.closest('[data-action]');
  if (actionButton) {
    const action = actionButton.dataset.action;
    if (action === 'fullscreen') {
      openFullscreen(key, channel);
      return;
    }
    if (action === 'stream') {
      const stream = state.hubs.get(key)?.streams.get(channel);
      if (stream) stopChannel(key, channel);
      else startChannel(key, channel);
      return;
    }
  }
  if (event.target.closest('select, input, label, button')) return;
  openDetails(key, channel);
});

ui.wallSection.addEventListener('change', event => {
  const tile = event.target.closest('.camera-tile');
  if (!tile) return;
  const key = tile.dataset.hubKey;
  const channel = Number(tile.dataset.channel);
  if (event.target.matches('.quality')) {
    setQuality(key, channel, event.target.value);
  }
  if (event.target.matches('.motion-toggle')) {
    toggleMotion(key, channel, event.target);
  }
});

ui.wallSection.addEventListener('keydown', event => {
  if (event.key !== 'Enter' && event.key !== ' ') return;
  const tile = event.target.closest('.camera-tile');
  if (!tile || event.target.closest('button, select, input, label')) return;
  event.preventDefault();
  openDetails(tile.dataset.hubKey, Number(tile.dataset.channel));
});

document.addEventListener('click', event => {
  const close = event.target.closest('[data-close]');
  if (!close) return;
  const dialog = close.closest('dialog');
  if (dialog) dialog.close();
});

wireDialogs();

document.addEventListener('fullscreenchange', () => {
  for (const video of document.querySelectorAll('video.stream-video')) {
    video.controls = document.fullscreenElement === video;
  }
});

document.addEventListener('i18n:change', () => {
  state.lang = i18n.getLanguage();
  resetForgetConfirm();
  if (streamCount()) updateStreamNotice();
  renderNotice();
  if (!ui.setupSection.hidden) {
    renderHubs();
    if (state.devices.length) renderDiagnostics();
  }
  if (!ui.wallSection.hidden) renderWall();
  if (ui.hubsDialog.open) renderHubManager();
  if (ui.settingsDialog.open) renderSettings();
  if (state.details) renderDetailsDrawer();
});

ui.forgetCode.addEventListener('click', () => forgetAllCredentials(ui.forgetCode));
ui.forgetAllCodes.addEventListener('click', () => forgetAllCredentials(ui.forgetAllCodes));

// Removing every code is irreversible for the session, so the first click only
// arms the action (ADR-0028); the second click within four seconds removes.
let forgetConfirmButton = null;
let forgetConfirmTimer = 0;

function resetForgetConfirm() {
  window.clearTimeout(forgetConfirmTimer);
  if (forgetConfirmButton) {
    i18n.setText(forgetConfirmButton, 'diagnostics.forgetCode');
    forgetConfirmButton = null;
  }
}

function forgetAllCredentials(button) {
  if (!button || forgetConfirmButton !== button) {
    resetForgetConfirm();
    forgetConfirmButton = button;
    i18n.setText(button, 'diagnostics.forgetCodeConfirm');
    forgetConfirmTimer = window.setTimeout(resetForgetConfirm, 4000);
    return;
  }
  resetForgetConfirm();
  state.credentials.clear();
  updateCredentialBadges();
  if (ui.hubsDialog.open) renderHubManager();
  setNotice('notice.forgotAllCodes');
}

ui.noticeAction.addEventListener('click', () => {
  const retry = state.noticeRetry;
  if (typeof retry !== 'function') return;
  state.noticeRetry = null;
  ui.noticeAction.hidden = true;
  retry();
});

ui.manualFallback.addEventListener('click', event => {
  const button = event.target.closest('[data-manual-channel]');
  if (!button || !state.manualHub) return;
  if (streamCount()) {
    setNotice('notice.stopBeforeReload', {}, 'error');
    return;
  }
  const key = hubKey(state.manualHub);
  let entry = state.hubs.get(key);
  if (!entry) {
    entry = { key, hub: state.manualHub, info: {}, channels: [], streams: new Map() };
    state.hubs.set(key, entry);
  }
  const index = Number(button.dataset.manualChannel);
  if (!entry.channels.some(channel => channel.index === index)) {
    entry.channels.push({ index, name: '', quality: '1' });
    entry.channels.sort((a, b) => a.index - b.index);
    state.filter = 'all';
    state.setupMode = false;
    renderLayout();
    renderHubs();
  }
});

window.addEventListener('beforeunload', () => {
  for (const hub of state.hubs.values()) {
    for (const stream of hub.streams.values()) {
      if (!stream.id) continue;
      navigator.sendBeacon('/api/live/stop', new Blob([JSON.stringify({id: stream.id})], {type:'application/json'}));
    }
  }
});

function openMobileMenu() {
  renderFilters();
  renderViewButtons();
  updateGlobalControls();
  if (!ui.mobileWallMenu.open) ui.mobileWallMenu.showModal();
}

/* -------------------------------------------------------------------------
 * Initialization
 * ---------------------------------------------------------------------- */

applyTheme(storedTheme());
ui.languageSelect.value = state.lang;
renderNotice();
renderLayout();
renderSettings();
syncMetricsVisibility();

if ('ResizeObserver' in window) {
  const observer = new ResizeObserver(scheduleLayout);
  observer.observe(ui.wallCanvas);
} else {
  window.addEventListener('resize', scheduleLayout);
}

document.fonts?.ready.then(scheduleLayout);

window.addEventListener('orientationchange', () => {
  setTimeout(scheduleLayout, 120);
});

i18n.ready.then(() => {
  state.lang = i18n.getLanguage();
  renderNotice();
  renderLayout();
  renderSettings();
  if (ui.hubsDialog.open) renderHubManager();
  if (state.details) renderDetailsDrawer();
});