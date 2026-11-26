import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

const codes: Record<string, string> = {
  'CMAG-TEST-AAA': 'code-a',
  'CMAG-TEST-BBB': 'code-b',
};

const defaultChannels: Record<string, Array<{ index: number; name: string }>> = {
  'CMAG-TEST-AAA': [{ index: 0, name: 'Terrace' }, { index: 1, name: 'Carport' }],
  'CMAG-TEST-BBB': [{ index: 0, name: 'Garden' }],
};

const playlist = '#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:2\n#EXT-X-ENDLIST\n';

type StreamStart = { did: string; cameraIndex: number };

async function mockGateway(page: Page, channelMap = defaultChannels) {
  const starts: StreamStart[] = [];
  const stops: string[] = [];

  await page.route(/\/api\/discover/, (route) => route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({
      devices: [
        { did: 'CMAG-TEST-AAA', ip: '192.0.2.10', model: 'SMH-T1R' },
        { did: 'CMAG-TEST-BBB', ip: '192.0.2.11', model: 'SMH-T1R' },
      ],
      durationMs: 1,
    }),
  }));

  await page.route(/\/api\/channels/, (route) => {
    const body = JSON.parse(route.request().postData() || '{}');
    if (codes[body.did] !== body.securityCode) {
      return route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ error: 'unauthorized', code: 'unauthorized' }) });
    }
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ channels: channelMap[body.did] || [] }),
    });
  });

  await page.route(/\/api\/live\/start/, (route) => {
    const body = JSON.parse(route.request().postData() || '{}');
    const id = `${body.did}-ch${body.cameraIndex}`;
    starts.push({ did: body.did, cameraIndex: body.cameraIndex });
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ id, playlistUrl: `/hls/${id}/index.m3u8` }),
    });
  });

  await page.route(/\/api\/live\/status/, (route) => route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ running: true, bitrateBytesPerSecond: 525000, framesPerSecond: 25, battery: { capacity: 87, charging: false }, uptimeSeconds: 11520 }),
  }));

  await page.route(/\/api\/live\/stop/, (route) => {
    const body = JSON.parse(route.request().postData() || '{}');
    stops.push(body.id);
    return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ stopped: true }) });
  });

  await page.route(/\/hls\//, (route) => route.fulfill({
    status: 200,
    contentType: 'application/vnd.apple.mpegurl',
    body: playlist,
  }));

  return { starts, stops };
}

async function discover(page: Page) {
  await page.getByRole('button', { name: 'Discover base stations' }).click();
  await expect(page.locator('.hub-card')).toHaveCount(2);
}

async function loadHub(page: Page, did: string) {
  await page.locator(`.load-channels[data-did="${did}"]`).click();
  await page.locator('#security-code').fill(codes[did]);
  await page.locator('#code-submit').click();
  await expect(page.locator('#wall-section')).toBeVisible();
}

async function loadBothHubs(page: Page) {
  await page.goto('/?lang=en');
  await discover(page);
  await loadHub(page, 'CMAG-TEST-AAA');
  await page.locator('#manage-hubs').click();
  await page.locator('#add-hub').click();
  await expect(page.locator('#setup-section')).toBeVisible();
  await expect(page.locator('.hub-card')).toHaveCount(2);
  await loadHub(page, 'CMAG-TEST-BBB');
}

async function startAll(page: Page, count: number) {
  await page.locator('#start-all-hubs').click();
  await expect(page.locator('.camera-tile .state-pill.live')).toHaveCount(count);
}

test.describe('live wall', () => {
  test.use({ colorScheme: 'light' });

  test('uses a full-width setup bar and empty state', async ({ page }) => {
    // Arrange
    await page.setViewportSize({ width: 2000, height: 1000 });
    await mockGateway(page);
    await page.goto('/?lang=en');

    // Act
    const widths = await page.evaluate(() => ({
      viewport: window.innerWidth,
      notice: document.getElementById('notice').getBoundingClientRect().width,
      setup: document.getElementById('setup-section').getBoundingClientRect().width,
      toolbar: document.querySelector('.setup-toolbar').getBoundingClientRect().width,
    }));

    // Assert
    expect(widths.notice).toBe(widths.viewport);
    expect(widths.setup).toBe(widths.viewport);
    expect(widths.toolbar).toBe(widths.viewport);
    await expect(page.locator('#setup-empty')).toBeVisible();
    await expect(page.getByRole('heading', { level: 2, name: 'No base station loaded' })).toBeVisible();
    await expect(page.locator('#hubs-section')).toBeHidden();
    await expect(page.locator('#discover')).toBeVisible();
    await expect(page.locator('.page-footer')).toHaveCount(0);

    // Act
    await page.getByRole('button', { name: 'Discover base stations' }).click();

    // Assert
    await expect(page.locator('.hub-card')).toHaveCount(2);
    await expect(page.locator('#setup-empty')).toBeHidden();
    await expect(page.locator('#hubs-section')).toBeVisible();
    await expect(page.getByRole('heading', { level: 2, name: 'Base stations' })).toBeAttached();
  });

  test('tiles are labelled groups with keyboard access', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');

    // Act
    const tile = page.locator('.camera-tile').first();
    await tile.focus();
    await page.keyboard.press('Enter');

    // Assert
    await expect(tile).toHaveAttribute('role', 'group');
    await expect(tile).toHaveAttribute('aria-label', /Terrace/);
    await expect(page.locator('#details-dialog')).toHaveJSProperty('open', true);
  });

  test('closes a dialog on a backdrop click', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');
    await page.locator('#wall-settings').click();
    const dialog = page.locator('#settings-dialog');
    await expect(dialog).toHaveJSProperty('open', true);

    // Act
    await page.mouse.click(5, 5);

    // Assert
    await expect(dialog).toHaveJSProperty('open', false);
  });

  test('keeps tiles and controls unclipped at 200% text scale', async ({ page }) => {
    // Arrange
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.addInitScript(() => { document.documentElement.style.fontSize = '32px'; });
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');
    await startAll(page, 2);
    await page.locator('.view-switch button[data-view="comfort"]').first().click();
    await page.waitForTimeout(300);

    // Assert: no horizontal page overflow
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
    expect(overflow).toBeLessThanOrEqual(1);

    // Assert: head and foot stay inside the tile
    const clipped = await page.evaluate(() => {
      const tile = document.querySelector('.camera-tile') as HTMLElement;
      const head = tile.querySelector('.tile-head') as HTMLElement;
      const foot = tile.querySelector('.tile-foot') as HTMLElement;
      const tileBox = tile.getBoundingClientRect();
      return {
        head: head.getBoundingClientRect().top < tileBox.top - 1,
        foot: foot.getBoundingClientRect().bottom > tileBox.bottom + 1,
      };
    });
    expect(clipped.head).toBe(false);
    expect(clipped.foot).toBe(false);

    // Assert: the metrics pill grows with the text instead of clipping it
    const pillFits = await page.locator('.tile-metrics').first().evaluate((element) => element.scrollHeight <= element.clientHeight + 1);
    expect(pillFits).toBe(true);
  });

  test('replaces setup with the wall and starts tiles at the top', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);

    // Act
    await loadHub(page, 'CMAG-TEST-AAA');

    // Assert
    await expect(page.locator('#setup-section')).toBeHidden();
    await expect(page.locator('#wall-section')).toBeVisible();
    await expect(page).toHaveTitle('AkkuCam Gateway');
    expect(await page.evaluate(() => document.body.classList.contains('wall-active'))).toBe(true);
    await expect(page.locator('.page-footer')).toHaveCount(0);
    await expect(page.locator('.camera-tile')).toHaveCount(2);
    await expect(page.locator('#hub-filters .filter-chip')).toHaveCount(2);
    const positions = await page.evaluate(() => {
      const canvas = document.getElementById('wall-canvas').getBoundingClientRect();
      const tile = document.querySelector('.camera-tile').getBoundingClientRect();
      return { offset: Math.round(tile.top - canvas.top) };
    });
    expect(positions.offset).toBe(12);
  });

  test('keeps equal tile sizes with an odd camera count', async ({ page }) => {
    // Arrange
    await mockGateway(page, {
      ...defaultChannels,
      'CMAG-TEST-AAA': [
        { index: 0, name: 'One' },
        { index: 1, name: 'Two' },
        { index: 2, name: 'Three' },
      ],
    });
    await page.goto('/?lang=en');
    await discover(page);

    // Act
    await loadHub(page, 'CMAG-TEST-AAA');

    // Assert
    await expect(page.locator('.camera-tile')).toHaveCount(3);
    await expect(page.locator('.tile-empty')).toHaveCount(1);
    const sizes = await page.locator('.camera-tile').evaluateAll((tiles) => tiles.map((tile) => {
      const rect = tile.getBoundingClientRect();
      return { width: Math.round(rect.width), height: Math.round(rect.height) };
    }));
    expect(new Set(sizes.map((size) => `${size.width}x${size.height}`)).size).toBe(1);
  });

  test('keeps the same video element when switching views', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');
    await startAll(page, 2);
    await page.evaluate(() => { document.querySelector('.camera-tile video.stream-video').dataset.marker = 'same'; });

    // Act and Assert
    await page.locator('[data-view="focus"]').first().click();
    await expect(page.locator('#focus-layout')).toBeVisible();
    await expect(page.locator('.camera-tile video.stream-video')).toHaveCount(1);
    expect(await page.evaluate(() => document.querySelector('.camera-tile video.stream-video')?.dataset.marker)).toBe('same');

    await page.locator('[data-view="comfort"]').first().click();
    await expect(page.locator('.camera-tile.is-comfort')).toHaveCount(2);
    await expect(page.locator('.camera-tile video.stream-video')).toHaveCount(2);
    expect(await page.evaluate(() => document.querySelector('.camera-tile video.stream-video')?.dataset.marker)).toBe('same');
  });

  test('shows live thumbnails in the focus strip', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');
    await startAll(page, 2);

    // Act
    await page.locator('[data-view="focus"]').first().click();

    // Assert
    await expect(page.locator('.focus-item .focus-thumb-canvas')).toHaveCount(2);
    await expect(page.locator('.focus-item .focus-thumb-canvas').first()).toBeVisible();
    await expect(page.locator('.focus-item .focus-thumb-placeholder').first()).toBeHidden();
  });

  test('shows live metrics on the tile without swallowing the click', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');
    await startAll(page, 2);
    const tile = page.locator('.camera-tile').first();
    const overlay = tile.locator('.tile-metrics');

    // Assert: KB/s instead of the coarse Mbit/s, plus fps and battery
    await expect(overlay).toBeVisible();
    await expect(overlay).toContainText('KB/s');
    await expect(overlay).toContainText('25 fps');
    await expect(overlay).toContainText('87 %');
    await expect(overlay).not.toContainText('Mbit/s');

    // Act: click the bottom-left corner where the overlay sits
    const box = (await tile.boundingBox())!;
    await page.mouse.click(box.x + 30, box.y + box.height - 22);

    // Assert: the overlay is decorative and the drawer opens
    await expect(page.locator('#details-dialog')).toHaveJSProperty('open', true);
    await expect(page.locator('#details-dialog .detail-grid').last()).toContainText('KB/s');
  });

  test('hides the hub label inside the video in auto and focus', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');

    // Assert: auto has no hub badge and the empty metrics pill stays hidden
    await expect(page.locator('.camera-tile .hub-badge')).toHaveCount(0);
    await expect(page.locator('.camera-tile .tile-metrics').first()).toBeHidden();

    // Assert: comfort keeps the hub under the camera name
    await page.locator('.view-switch button[data-view="comfort"]').first().click();
    await expect(page.locator('.camera-tile .tile-titles span').first()).toContainText('SMH-T1R');

    // Assert: focus has neither the badge nor the strip label
    await page.locator('.view-switch button[data-view="focus"]').first().click();
    await expect(page.locator('.camera-tile .hub-badge')).toHaveCount(0);
    await expect(page.locator('.focus-item small')).toHaveCount(0);
  });

  test('toggles the live metrics overlay in the wall settings', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');
    await startAll(page, 2);
    const overlay = page.locator('.camera-tile .tile-metrics').first();
    await expect(overlay).toBeVisible();

    // Act: switch the overlay off
    await page.locator('#wall-settings').click();
    const toggle = page.locator('#settings-live-metrics');
    await expect(toggle).toBeChecked();
    await toggle.uncheck();
    await expect(page.locator('body')).toHaveClass(/metrics-off/);
    expect(await page.evaluate(() => localStorage.getItem('akkucam-wall-metrics'))).toBe('off');
    await page.locator('#settings-dialog [data-close]').last().click();

    // Assert
    await expect(overlay).toBeHidden();

    // Act: reset restores the default
    await page.locator('#wall-settings').click();
    await page.locator('#reset-wall-settings').click();
    await expect(toggle).toBeChecked();
    expect(await page.evaluate(() => localStorage.getItem('akkucam-wall-metrics'))).toBeNull();
    await page.locator('#settings-dialog [data-close]').last().click();
    await expect(overlay).toBeVisible();
  });

  test('starts and stops from the details drawer', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');
    const tile = page.locator('.camera-tile').first();
    const details = page.locator('#details-dialog');
    await tile.click();
    await expect(details).toBeVisible();

    // Act
    await details.getByRole('button', { name: 'Start', exact: true }).click();

    // Assert
    await expect(tile.locator('.state-pill.live')).toBeVisible();
    await expect(details.getByRole('button', { name: 'Stop', exact: true })).toBeVisible();

    // Act
    await details.getByRole('button', { name: 'Stop', exact: true }).click();

    // Assert
    await expect(tile.locator('.state-pill')).toContainText('Ready');
    await expect(details.getByRole('button', { name: 'Start', exact: true })).toBeVisible();
  });

  test('starts and stops from the comfort footer', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');
    await page.locator('[data-view="comfort"]').first().click();
    const tile = page.locator('.camera-tile.is-comfort').first();
    const button = tile.locator('.stream-toggle');
    await expect(button).toHaveText('Start');

    // Act
    await button.click();

    // Assert
    await expect(tile.locator('.state-pill.live')).toBeVisible();
    await expect(tile.locator('.stream-toggle')).toHaveText('Stop');

    // Act
    await tile.locator('.stream-toggle').click();

    // Assert
    await expect(tile.locator('.state-pill')).toContainText('Ready');
    await expect(tile.locator('.stream-toggle')).toHaveText('Start');
  });

  test('starts and stops one hub without affecting the other', async ({ page }) => {
    // Arrange
    const { starts } = await mockGateway(page);
    await loadBothHubs(page);
    await page.locator('#manage-hubs').click();
    const rowA = page.locator('#hub-manager .hub-row[data-hub-key="CMAG-TEST-AAA"]');

    // Act
    await rowA.getByRole('button', { name: 'Start all' }).click();

    // Assert
    await expect(page.locator('.camera-tile[data-hub-key="CMAG-TEST-AAA"] .state-pill.live')).toHaveCount(2);
    await expect(page.locator('.camera-tile[data-hub-key="CMAG-TEST-BBB"] .state-pill.live')).toHaveCount(0);
    expect(starts.map((item) => item.did)).toEqual(['CMAG-TEST-AAA', 'CMAG-TEST-AAA']);

    // Act
    await rowA.getByRole('button', { name: 'Stop all' }).click();

    // Assert
    await expect(page.locator('.camera-tile[data-hub-key="CMAG-TEST-AAA"] .state-pill.live')).toHaveCount(0);
    await expect(page.locator('.camera-tile[data-hub-key="CMAG-TEST-BBB"] .state-pill')).toContainText('Ready');
  });

  test('filters base stations without stopping their streams', async ({ page }) => {
    // Arrange
    const { starts, stops } = await mockGateway(page);
    await loadBothHubs(page);
    await startAll(page, 3);

    // Act
    await page.locator('#hub-filters [data-filter="CMAG-TEST-BBB"]').click();

    // Assert
    await expect(page.locator('.camera-tile')).toHaveCount(1);
    await expect(page.locator('.camera-tile .state-pill.live')).toHaveCount(1);
    expect(starts).toHaveLength(3);
    expect(stops).toHaveLength(0);

    // Act
    await page.locator('#hub-filters [data-filter="all"]').click();

    // Assert
    await expect(page.locator('.camera-tile')).toHaveCount(3);
    await expect(page.locator('.camera-tile .state-pill.live')).toHaveCount(3);
  });

  test('persists the selected view and resets it', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');
    await page.locator('#wall-settings').click();
    await page.check('input[name="wall-view"][value="focus"]');
    await expect(page.locator('#focus-layout')).toBeVisible();
    await expect(page.locator('#settings-dialog a')).toHaveCount(0);
    await page.locator('#settings-dialog .icon-button[data-close]').click();
    expect(await page.evaluate(() => localStorage.getItem('akkucam-wall-view'))).toBe('focus');

    // Act
    await page.reload();
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');

    // Assert
    await expect(page.locator('#focus-layout')).toBeVisible();
    await expect(page.locator('body')).toHaveAttribute('data-view', 'focus');

    // Act
    await page.locator('#wall-settings').click();
    await page.locator('#reset-wall-settings').click();

    // Assert
    await expect(page.locator('#wall-grid')).toBeVisible();
    await expect(page.locator('body')).toHaveAttribute('data-view', 'auto');
    expect(await page.evaluate(() => localStorage.getItem('akkucam-wall-view'))).toBeNull();
  });

  test('uses the compact wall bar on a phone', async ({ page }) => {
    // Arrange
    await page.setViewportSize({ width: 390, height: 844 });
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);

    // Act
    await loadHub(page, 'CMAG-TEST-AAA');

    // Assert
    await expect(page.locator('.wall-toolbar')).toBeHidden();
    await expect(page.locator('#mobile-wall-bar')).toBeVisible();
    await expect(page.locator('#mobile-filter-label')).toHaveText('All');
    await expect(page.locator('#mobile-filter-count')).toHaveText('2');

    // Act
    await page.locator('#mobile-menu-button').click();
    await expect(page.locator('#mobile-wall-menu')).toBeVisible();
    await page.locator('#mobile-wall-menu [data-view="comfort"]').click();

    // Assert
    await expect(page.locator('#mobile-wall-menu')).toBeHidden();
    await expect(page.locator('.camera-tile.is-comfort')).toHaveCount(2);
    await expect(page.locator('body')).toHaveAttribute('data-view', 'comfort');

    // Act
    await page.locator('#mobile-menu-button').click();
    await page.locator('#mobile-settings').click();

    // Assert
    await expect(page.locator('#settings-dialog')).toBeVisible();
  });

  test('limits the compact wall bar to phone sizes', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');

    // Act: iPhone landscape.
    await page.setViewportSize({ width: 932, height: 430 });

    // Assert
    await expect(page.locator('.wall-toolbar')).toBeHidden();
    await expect(page.locator('#mobile-wall-bar')).toBeVisible();

    // Act: iPad portrait.
    await page.setViewportSize({ width: 820, height: 1180 });

    // Assert
    await expect(page.locator('.wall-toolbar')).toBeVisible();
    await expect(page.locator('#mobile-wall-bar')).toBeHidden();
  });

  test('updates labels while a stream keeps running', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await page.goto('/?lang=en');
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA');
    await startAll(page, 2);
    await page.evaluate(() => { document.querySelector('.camera-tile video.stream-video').dataset.marker = 'same'; });

    // Act
    await page.locator('#language-select').selectOption('de');

    // Assert
    await expect(page.locator('.camera-tile .state-pill.live')).toHaveCount(2);
    await expect(page.locator('#hub-filters .filter-chip').first()).toContainText('Alle');
    await expect(page.locator('#layout-status')).toHaveCount(0);
    expect(await page.evaluate(() => document.querySelector('.camera-tile video.stream-video')?.dataset.marker)).toBe('same');
  });
});