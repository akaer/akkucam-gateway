import { test, expect } from '@playwright/test';

const devices = [
  {
    ip: '192.0.2.10',
    did: 'CMAG-000000-TEST1',
    model: 'SMH-T1R',
    local: true,
  },
];

async function stubDiscovery(page, body = { devices, durationMs: 100 }) {
  await page.route('**/api/discover', (route) => route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify(body),
  }));
}

test.describe('language detection', () => {
  test.use({ locale: 'de-DE', colorScheme: 'light' });

  test('German browser starts in German', async ({ page }) => {
    // Arrange
    await page.goto('/');

    // Act
    const discover = page.locator('#discover');

    // Assert
    await expect(discover).toContainText('Basisstationen suchen');
    await expect(page.locator('html')).toHaveAttribute('lang', 'de');
    await expect(page.locator('#language-select')).toHaveValue('de');
  });

  test('plurals follow the active language', async ({ page }) => {
    // Arrange
    await stubDiscovery(page);
    await page.goto('/');

    // Act
    await page.getByRole('button', { name: 'Basisstationen suchen' }).click();

    // Assert
    await expect(page.locator('#notice')).toContainText('1 Gerät in 0,1 s gefunden.');
  });
});

test.describe('unsupported browser language', () => {
  test.use({ locale: 'fr-FR', colorScheme: 'light' });

  test('falls back to English', async ({ page }) => {
    // Arrange
    await page.goto('/');

    // Act
    const discover = page.locator('#discover');

    // Assert
    await expect(discover).toContainText('Discover base stations');
    await expect(page.locator('html')).toHaveAttribute('lang', 'en');
    await expect(page.locator('#language-select')).toHaveValue('en');
  });
});

test.describe('query override', () => {
  test.use({ locale: 'en-US', colorScheme: 'light' });

  test('?lang=de overrides the English browser', async ({ page }) => {
    // Arrange
    await page.goto('/?lang=de');

    // Act
    const discover = page.locator('#discover');

    // Assert
    await expect(discover).toContainText('Basisstationen suchen');
    await expect(page.locator('html')).toHaveAttribute('lang', 'de');
  });

  test('?lang is not persisted', async ({ page }) => {
    // Arrange
    await page.goto('/?lang=de');
    await expect(page.locator('html')).toHaveAttribute('lang', 'de');

    // Act
    const stored = await page.evaluate(() => localStorage.getItem('akkucam-lang'));

    // Assert
    expect(stored).toBeNull();
  });
});

test.describe('query override from German', () => {
  test.use({ locale: 'de-DE', colorScheme: 'light' });

  test('?lang=en overrides the German browser', async ({ page }) => {
    // Arrange
    await page.goto('/?lang=en');

    // Act
    const discover = page.locator('#discover');

    // Assert
    await expect(discover).toContainText('Discover base stations');
    await expect(page.locator('html')).toHaveAttribute('lang', 'en');
  });
});

test.describe('live switching', () => {
  test.use({ locale: 'en-US', colorScheme: 'light' });

  test('switching language does not reload the page', async ({ page }) => {
    // Arrange
    await page.goto('/');
    await expect(page.locator('#discover')).toContainText('Discover base stations');
    await page.evaluate(() => {
      window.__marker = 'alive';
      window.__navigations = performance.getEntriesByType('navigation').length;
    });

    // Act
    await page.locator('#language-select').selectOption('de');

    // Assert
    await expect(page.locator('#discover')).toContainText('Basisstationen suchen');
    expect(await page.evaluate(() => window.__marker)).toBe('alive');
    expect(await page.evaluate(() => performance.getEntriesByType('navigation').length)).toBe(1);
  });

  test('language choice persists across reload', async ({ page }) => {
    // Arrange
    await page.goto('/');
    await page.locator('#language-select').selectOption('de');
    await expect(page.locator('#discover')).toContainText('Basisstationen suchen');

    // Act
    await page.reload();

    // Assert
    await expect(page.locator('#discover')).toContainText('Basisstationen suchen');
    await expect(page.locator('html')).toHaveAttribute('lang', 'de');
    expect(await page.evaluate(() => localStorage.getItem('akkucam-lang'))).toBe('de');
  });

  test('keyboard changes the language', async ({ page }) => {
    // Arrange
    await page.goto('/');
    const select = page.locator('#language-select');
    await select.focus();
    expect(await page.evaluate(() => document.activeElement?.id)).toBe('language-select');

    // Act
    await page.keyboard.press('ArrowDown');

    // Assert
    await expect(select).toHaveValue('de');
    await expect(page.locator('html')).toHaveAttribute('lang', 'de');
  });

  test('html lang and aria labels follow the language', async ({ page }) => {
    // Arrange
    await page.goto('/');

    // Act
    await page.locator('#language-select').selectOption('de');

    // Assert
    await expect(page.locator('html')).toHaveAttribute('lang', 'de');
    await expect(page.locator('#language-select')).toHaveAttribute('aria-label', 'Sprache');
    await expect(page.locator('#theme-toggle')).toHaveAttribute('aria-label', 'Zum hellen Farbschema wechseln');
    await expect(page.getByRole('button', { name: 'Basisstationen suchen' })).toBeVisible();
  });
});

test.describe('help popup', () => {
  test.use({ colorScheme: 'light' });

  test('opens with the localized help text', async ({ page }) => {
    // Arrange
    await page.goto('/?lang=de');

    // Act
    await page.locator('#help-toggle').click();

    // Assert
    const dialog = page.locator('#help-dialog');
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText('So funktioniert es');
    await expect(dialog).toContainText('Basisstation finden, gekoppelte Kameras öffnen und Livebild ohne Cloud-Dienst starten.');
    await expect(dialog.getByRole('link', { name: 'Projekt auf GitHub' })).toHaveAttribute('href', 'https://github.com/akaer/akkucam-gateway');
  });

  test('links to the project repository', async ({ page }) => {
    // Arrange
    await page.goto('/?lang=en');
    await page.locator('#help-toggle').click();

    // Act
    const link = page.locator('#help-dialog').getByRole('link', { name: 'Project on GitHub' });

    // Assert
    await expect(link).toBeVisible();
    await expect(link).toHaveAttribute('href', 'https://github.com/akaer/akkucam-gateway');
    await expect(link).toHaveAttribute('target', '_blank');
    await expect(link).toHaveAttribute('rel', 'noopener');
  });

  test('shows the version and build commit from the API', async ({ page, request }) => {
    // Arrange
    const response = await request.get('/api/version');
    const info = await response.json();

    // Act
    await page.goto('/?lang=en');
    await page.locator('#help-toggle').click();

    // Assert
    const version = page.locator('#app-version');
    await expect(version).toBeVisible();
    await expect(version).toContainText(info.version);
    if (info.commit) {
      await expect(version).toContainText(info.commit);
    }
  });

  test('help icon sits at the far right of the topbar', async ({ page }) => {
    // Arrange
    await page.goto('/?lang=en');

    // Act
    const positions = await page.evaluate(() => {
      const rect = (selector) => document.querySelector(selector).getBoundingClientRect();
      return { help: rect('#help-toggle').right, theme: rect('#theme-toggle').right };
    });

    // Assert
    expect(positions.help).toBeGreaterThan(positions.theme);
  });

  test('dismiss button closes the popup', async ({ page }) => {
    // Arrange
    await page.goto('/?lang=en');
    await page.locator('#help-toggle').click();
    const dialog = page.locator('#help-dialog');
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText('Find a base station, open its cameras and start a live view without any cloud service.');

    // Act
    await dialog.getByRole('button', { name: 'Got it' }).click();

    // Assert
    await expect(dialog).toBeHidden();
  });
});

test.describe('error translation', () => {
  test.use({ locale: 'de-DE', colorScheme: 'light' });

  test('API error codes are translated', async ({ page }) => {
    // Arrange
    await page.route('**/api/discover', (route) => route.fulfill({
      status: 409,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'Stop the active streams first.', code: 'conflict' }),
    }));
    await page.goto('/');

    // Act
    await page.getByRole('button', { name: 'Basisstationen suchen' }).click();

    // Assert
    await expect(page.locator('#notice')).toContainText('Beende zuerst die aktiven Streams.');
    await expect(page.locator('#notice')).toHaveClass(/error/);
  });
});

test.describe('hub actions', () => {
  test.use({ colorScheme: 'light' });

  test('reload and remove actions use the active language', async ({ page }) => {
    // Arrange
    await page.route('**/api/discover', (route) => route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        devices: [{ ip: '192.0.2.20', did: 'CMAG-TEST-DE', model: 'SMH-T1R' }],
        durationMs: 1,
      }),
    }));
    await page.route('**/api/channels', (route) => route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ channels: [{ index: 1, name: 'Terrasse' }] }),
    }));
    await page.goto('/?lang=de');
    await page.getByRole('button', { name: 'Basisstationen suchen' }).click();

    // Act
    await page.locator('.load-channels').click();
    await page.locator('#security-code').fill('code-a');
    await page.locator('#code-submit').click();

    // Assert
    const card = page.locator('.hub-card');
    await expect(card.locator('.load-channels')).toHaveText('Kameras neu laden');
    await expect(card.locator('.forget-hub-code')).toHaveText('Code entfernen');
  });
});
