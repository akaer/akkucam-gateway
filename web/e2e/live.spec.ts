import { test, expect, type Page } from '@playwright/test';

const securityCode = process.env.AKKUCAM_TEST_SECURITY_CODE;
const hubDID = process.env.AKKUCAM_TEST_HUB_DID;

async function discoverAndLoadChannels(page: Page) {
  await page.goto('/');
  await page.getByRole('button', { name: 'Discover base stations' }).click();

  const hubCard = page.locator('.hub-card').filter({
    has: page.locator('.load-channels'),
  }).first();
  await expect(hubCard).toBeVisible({ timeout: 30_000 });

  if (hubDID) {
    const did = await hubCard.locator('.load-channels').getAttribute('data-did');
    expect(did).toBe(hubDID);
  }

  await hubCard.getByRole('button', { name: 'Load cameras' }).click();

  const dialog = page.locator('#code-dialog');
  await expect(dialog).toBeVisible();
  await page.locator('#security-code').fill(securityCode!);
  await page.locator('#code-submit').click();
  await expect(dialog).toBeHidden();

  const tiles = page.locator('.camera-tile');
  await expect(tiles.first()).toBeVisible({ timeout: 30_000 });
  expect(await tiles.count()).toBeGreaterThan(0);
}

test.describe('live gateway', () => {
  test.skip(!securityCode, 'AKKUCAM_TEST_SECURITY_CODE is not set');

  test('discovers a hub, authenticates and loads channels', async ({ page }) => {
    // Arrange
    await discoverAndLoadChannels(page);

    // Act
    await page.evaluate(() => {
      window.__marker = 'alive';
      window.__navigations = performance.getEntriesByType('navigation').length;
    });
    await page.locator('#language-select').selectOption('de');

    // Assert
    await expect(page.locator('#discover')).toContainText('Basisstationen suchen');
    await expect(page.locator('.camera-tile').first().locator('.state-pill')).toContainText('Bereit');
    expect(await page.evaluate(() => window.__marker)).toBe('alive');
    expect(await page.evaluate(() => performance.getEntriesByType('navigation').length)).toBe(1);

    const storage = await page.evaluate(() => ({
      local: Object.fromEntries(Object.keys(localStorage).map((key) => [key, localStorage.getItem(key)])),
      session: Object.fromEntries(Object.keys(sessionStorage).map((key) => [key, sessionStorage.getItem(key)])),
    }));
    expect(JSON.stringify(storage)).not.toContain(securityCode!);
  });

  test('starts and stops a live stream', async ({ page }) => {
    test.skip(process.env.AKKUCAM_TEST_LIVE_STREAM !== '1', 'set AKKUCAM_TEST_LIVE_STREAM=1 to wake the camera');

    // Arrange
    await discoverAndLoadChannels(page);
    const tile = page.locator('.camera-tile').first();
    const details = page.locator('#details-dialog');

    // Act
    await tile.click();
    await expect(details).toBeVisible();
    await details.getByRole('button', { name: 'Start', exact: true }).click();

    // Assert
    await expect(tile.locator('.state-pill.live')).toBeVisible({ timeout: 60_000 });
    await expect(tile.locator('video')).toBeVisible();
    await details.locator('[data-close]').click();

    // Act: switch language while the stream is running
    await page.evaluate(() => {
      document.querySelector('.camera-tile video.stream-video')!.dataset.marker = 'same';
    });
    await page.locator('#language-select').selectOption('de');

    // Assert: the same video element is still mounted and live
    await expect(tile.locator('.state-pill.live')).toBeVisible();
    expect(await page.evaluate(() => document.querySelector('.camera-tile video.stream-video')?.dataset.marker)).toBe('same');

    // Act: switch back to English and stop the stream
    await page.locator('#language-select').selectOption('en');
    await tile.click();
    await expect(details).toBeVisible();
    await details.getByRole('button', { name: 'Stop', exact: true }).click();

    // Assert
    await expect(tile.locator('.state-pill')).toContainText('Ready');
    await expect(tile.locator('.state-pill.live')).toHaveCount(0);
  });
});
