import { test, expect } from '@playwright/test';


// WCAG relative luminance for an sRGB colour in rgb()/rgba() notation.
function channels(value: string): [number, number, number] {
  const parts = (value.match(/[0-9.]+/g) || []).map(Number);
  return [parts[0], parts[1], parts[2]];
}

function luminance([r, g, b]: [number, number, number]): number {
  const linear = (channel: number) => {
    const value = channel / 255;
    return value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  };
  return 0.2126 * linear(r) + 0.7152 * linear(g) + 0.0722 * linear(b);
}

function contrast(foreground: [number, number, number], background: [number, number, number]): number {
  const a = luminance(foreground);
  const b = luminance(background);
  return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);
}

test.describe('branding and licensing', () => {
  test.use({ colorScheme: 'light' });

  test('brand mark renders the camera icon', async ({ page }) => {
    // Arrange
    await page.goto('/');

    // Act
    const icon = page.locator('.brand-mark svg');

    // Assert: the brand uses the shared sprite (ADR-0029), not a second copy.
    await expect(icon).toBeVisible();
    await expect(icon.locator('use')).toHaveAttribute('href', '#icon-camera');
    await expect(page.locator('symbol#icon-camera')).toHaveAttribute('viewBox', '0 0 512 512');
    await expect(page.locator('symbol#icon-camera path')).toHaveCount(1);
  });

  for (const scheme of ['light', 'dark'] as const) {
    test(`primary text meets WCAG AA in the ${scheme} theme`, async ({ page }) => {
      // Arrange
      await page.emulateMedia({ colorScheme: scheme });
      await page.goto('/');

      // Act
      const colours = await page.locator('#discover').evaluate((element) => {
        const style = getComputedStyle(element);
        return { foreground: style.color, background: style.backgroundColor };
      });
      const notice = await page.locator('#notice').evaluate((element) => ({
        foreground: getComputedStyle(element.querySelector('.notice-text') as Element).color,
        background: getComputedStyle(element).backgroundColor,
      }));

      // Assert
      expect(contrast(channels(colours.foreground), channels(colours.background))).toBeGreaterThanOrEqual(4.5);
      expect(contrast(channels(notice.foreground), channels(notice.background))).toBeGreaterThanOrEqual(4.5);
    });
  }

  test('uses only the shipped font weights', async ({ page }) => {
    // Arrange
    await page.goto('/');

    // Act
    const weights = await page.evaluate(() => {
      const selectors = ['body', 'h2', '.brand strong', '.notice-text', 'button.primary'];
      return selectors.map((selector) => [selector, getComputedStyle(document.querySelector(selector) as Element).fontWeight]);
    });

    // Assert
    for (const [selector, weight] of weights) {
      expect(['400', '700'], selector).toContain(weight);
    }
  });

  test('theme toggle returns to the system preference', async ({ page }) => {
    // Arrange: system prefers dark, nothing stored yet
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.goto('/');

    // Assert: following the system means no data-theme and a dark canvas
    await expect(page.locator('html')).not.toHaveAttribute('data-theme', /.*/);
    const initial = await page.locator('body').evaluate((element) => getComputedStyle(element).backgroundColor);
    expect(initial).toBe('rgb(46, 52, 64)');

    // Act: first click pins the opposite scheme
    await page.locator('#theme-toggle').click();
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
    expect(await page.evaluate(() => localStorage.getItem('akkucam-theme'))).toBe('light');

    // Act: second click matches the system again and drops the override
    await page.locator('#theme-toggle').click();
    await expect(page.locator('html')).not.toHaveAttribute('data-theme', /.*/);
    expect(await page.evaluate(() => localStorage.getItem('akkucam-theme'))).toBeNull();
    const restored = await page.locator('body').evaluate((element) => getComputedStyle(element).backgroundColor);
    expect(restored).toBe('rgb(46, 52, 64)');
  });

  test('favicon follows the dashboard theme', async ({ page }) => {
    // Arrange
    await page.goto('/');
    const favicon = page.locator('link#favicon');
    await expect(favicon).toHaveAttribute('href', '/icon-light.svg');

    // Act
    await page.locator('#theme-toggle').click();

    // Assert
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
    await expect(favicon).toHaveAttribute('href', '/icon-dark.svg');
  });

  test('dashboard has no footer or copyright line', async ({ page }) => {
    // Arrange
    await page.goto('/');

    // Act
    const body = await page.locator('body').innerText();

    // Assert
    await expect(page.locator('.page-footer')).toHaveCount(0);
    expect(body).not.toContain('© 2026 André Raabe');
    expect(body).not.toContain('MIT-Lizenz');
  });

  test('help popup opens the scrollable license dialog', async ({ page }) => {
    // Arrange
    await page.goto('/?lang=en');
    await page.locator('#help-toggle').click();
    await expect(page.locator('#help-dialog')).toBeVisible();

    // Act
    await page.locator('#open-licenses').click();

    // Assert
    const dialog = page.locator('#licenses-dialog');
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText('Licenses and third-party components');
    await expect(dialog).toContainText('Project license: MIT');
    const licenses = dialog.locator('.license-list');
    await expect(licenses).toContainText('hls.js');
    await expect(licenses).toContainText('Apache-2.0');
    await expect(licenses).toContainText('Font Awesome Free');
    await expect(licenses).toContainText('CC BY 4.0');
    await expect(licenses).toContainText('Inter Tight');
    await expect(licenses).toContainText('OFL-1.1');
    await expect(licenses).toContainText('Swagger UI');
    await expect(licenses).toContainText('FFmpeg');
    await expect(licenses).toContainText('Alpine Linux');
    await expect(licenses).toContainText('Go');
    await expect(dialog.locator('a[href="/api/docs"]')).toBeVisible();
    expect(await dialog.locator('.license-scroll').evaluate((element) => getComputedStyle(element).overflowY)).toBe('auto');
    for (const link of await dialog.locator('.license-list a').all()) {
      const box = await link.boundingBox();
      expect(box?.height ?? 0).toBeGreaterThanOrEqual(24);
    }
    await expect(page.locator('#help-dialog')).toBeHidden();
  });

  test('reduced motion disables the chevron transition', async ({ page }) => {
    // Arrange
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await page.goto('/?lang=en');

    // Act
    const duration = await page.locator('.chevron').first().evaluate((element) => getComputedStyle(element).transitionDuration);

    // Assert
    expect(duration).toBe('0s');
  });

  test('license dialog is translated in German', async ({ page }) => {
    // Arrange
    await page.goto('/?lang=de');
    await page.locator('#help-toggle').click();

    // Act
    await page.locator('#open-licenses').click();

    // Assert
    const dialog = page.locator('#licenses-dialog');
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText('Lizenzen und Drittkomponenten');
    await expect(dialog).toContainText('Projektlizenz: MIT');
    await expect(dialog.locator('a[href="/api/docs"]')).toHaveText('API-Dokumentation');
  });

  test('brand name uses the vendored Inter Tight font', async ({ page }) => {
    // Arrange
    await page.goto('/');

    // Act
    await page.evaluate(() => document.fonts.ready);
    const brand = page.locator('.brand strong');
    const fontFamily = await brand.evaluate((element) => getComputedStyle(element).fontFamily);
    const fontWeight = await brand.evaluate((element) => getComputedStyle(element).fontWeight);
    const fontLoaded = await page.evaluate(() => document.fonts.check('700 16px "Inter Tight"'));

    // Assert
    expect(fontFamily).toContain('Inter Tight');
    expect(fontWeight).toBe('700');
    expect(fontLoaded).toBe(true);
  });

  test('setup and wall headings use the vendored Inter Tight font', async ({ page }) => {
    // Arrange
    await page.goto('/');
    await page.evaluate(() => document.fonts.ready);

    // Act and Assert
    for (const selector of ['.setup-toolbar h1', '#wall-heading']) {
      const fontFamily = await page.locator(selector).evaluate((element) => getComputedStyle(element).fontFamily);
      expect(fontFamily).toContain('Inter Tight');
    }
  });

  test('license files, fonts and theme icons are served', async ({ page }) => {
    // Arrange
    await page.goto('/');

    // Act and Assert
    for (const path of [
      '/LICENSE.txt',
      '/HLS-LICENSE.txt',
      '/FONTAWESOME-LICENSE.txt',
      '/INTERTIGHT-LICENSE.txt',
      '/swagger/LICENSE',
      '/fonts/inter-tight-latin.woff2',
      '/fonts/inter-tight-bold-latin.woff2',
      '/icon-light.svg',
      '/icon-dark.svg',
    ]) {
      const response = await page.request.get(path);
      expect(response.status(), path).toBe(200);
    }
  });
});