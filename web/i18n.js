// Minimal client-side i18n for the dashboard.
// Catalogs live in /locales/<lang>.json and are served by the same static file
// server as the rest of the dashboard. No build step, no external CDN.
(() => {
  const SUPPORTED = ['en', 'de'];
  const FALLBACK = 'en';

  const state = {
    lang: FALLBACK,
    catalogs: {},
    ready: Promise.resolve(),
  };

  function readStoredLanguage() {
    try {
      return localStorage.getItem('akkucam-lang');
    } catch {
      return null;
    }
  }

  function storeLanguage(lang) {
    try {
      localStorage.setItem('akkucam-lang', lang);
    } catch {
      // Storage may be unavailable (private mode); the language still applies
      // for the current page view.
    }
  }

  function browserPrefersGerman() {
    const languages = navigator.languages && navigator.languages.length
      ? navigator.languages
      : [navigator.language || ''];
    return languages.some((item) => String(item).toLowerCase().startsWith('de'));
  }

  function detectLanguage() {
    const query = new URLSearchParams(location.search).get('lang');
    if (query && SUPPORTED.includes(query.toLowerCase())) {
      return query.toLowerCase();
    }
    const stored = readStoredLanguage();
    if (stored && SUPPORTED.includes(stored)) {
      return stored;
    }
    return browserPrefersGerman() ? 'de' : FALLBACK;
  }

  function loadCatalog(lang) {
    return fetch(`/locales/${lang}.json`, { cache: 'no-cache' })
      .then((response) => (response.ok ? response.json() : null))
      .then((catalog) => {
        if (catalog && typeof catalog === 'object') {
          state.catalogs[lang] = catalog;
        }
      })
      .catch(() => {
        // A missing catalog must not break the page; the inline HTML text is
        // the English fallback.
      });
  }

  function readParams(element) {
    const raw = element.dataset.i18nParams;
    if (!raw) return undefined;
    try {
      return JSON.parse(raw);
    } catch {
      return undefined;
    }
  }

  function t(key, params) {
    const values = params || {};
    const active = state.catalogs[state.lang] || {};
    const fallback = state.catalogs[FALLBACK] || {};

    let template = active[key];
    if (template === undefined && values.count !== undefined) {
      const suffix = Number(values.count) === 1 ? '.one' : '.other';
      template = active[key + suffix] ?? fallback[key + suffix];
    }
    if (template === undefined) {
      template = fallback[key];
    }
    if (template === undefined) {
      return key;
    }
    return String(template).replace(/\{(\w+)\}/g, (_, name) => (
      values[name] === undefined ? '' : String(values[name])
    ));
  }

  function has(key) {
    return Boolean(
      (state.catalogs[state.lang] && state.catalogs[state.lang][key] !== undefined)
      || (state.catalogs[FALLBACK] && state.catalogs[FALLBACK][key] !== undefined),
    );
  }

  function setText(element, key, params) {
    if (!element) return;
    element.dataset.i18n = key;
    if (params) {
      element.dataset.i18nParams = JSON.stringify(params);
    } else {
      delete element.dataset.i18nParams;
    }
    element.textContent = t(key, params);
  }

  function applyTranslations(root) {
    const scope = root || document;
    scope.querySelectorAll('[data-i18n]').forEach((element) => {
      element.textContent = t(element.dataset.i18n, readParams(element));
    });
    scope.querySelectorAll('[data-i18n-title]').forEach((element) => {
      element.title = t(element.dataset.i18nTitle);
    });
    scope.querySelectorAll('[data-i18n-aria-label]').forEach((element) => {
      element.setAttribute('aria-label', t(element.dataset.i18nAriaLabel));
    });
    // The dashboard uses app.title; standalone pages can override the document
    // title with data-i18n-document-title on <html>.
    document.title = t(document.documentElement.dataset.i18nDocumentTitle || 'app.title');
    document.documentElement.lang = state.lang;
  }

  function setLanguage(lang, options) {
    const next = SUPPORTED.includes(lang) ? lang : FALLBACK;
    const persist = !options || options.persist !== false;
    state.lang = next;
    document.documentElement.lang = next;
    document.documentElement.dataset.lang = next;
    if (persist) storeLanguage(next);
    applyTranslations();
    document.dispatchEvent(new CustomEvent('i18n:change', { detail: { lang: next } }));
  }

  function getLanguage() {
    return state.lang;
  }

  state.lang = detectLanguage();
  document.documentElement.lang = state.lang;
  document.documentElement.dataset.lang = state.lang;

  state.ready = Promise.all([loadCatalog('en'), loadCatalog('de')]).then(() => {
    applyTranslations();
  });

  window.i18n = {
    t,
    has,
    setText,
    applyTranslations,
    setLanguage,
    getLanguage,
    ready: state.ready,
  };
})();
