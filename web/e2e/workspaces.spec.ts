import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

// Fake credentials only. Real codes must never be committed to the repository.
const codes: Record<string, string> = {
  'CMAG-TEST-AAA': 'code-a',
  'CMAG-TEST-BBB': 'code-b',
};

const channels: Record<string, Array<{ index: number; name: string }>> = {
  'CMAG-TEST-AAA': [{ index: 0, name: 'Terrace' }, { index: 1, name: 'Carport' }],
  'CMAG-TEST-BBB': [{ index: 0, name: 'Garden' }],
};

type StreamStart = { did: string; cameraIndex: number };

const playlist = '#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:2\n#EXT-X-ENDLIST\n';

async function mockGateway(page: Page) {
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
      return route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'unauthorized', code: 'unauthorized' }),
      });
    }
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ channels: channels[body.did] || [] }),
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
    body: JSON.stringify({ running: true, bitrateBytesPerSecond: 525000, framesPerSecond: 25 }),
  }));

  await page.route(/\/api\/live\/stop/, (route) => {
    const body = JSON.parse(route.request().postData() || '{}');
    stops.push(body.id);
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ stopped: true }),
    });
  });

  await page.route(/\/hls\//, (route) => route.fulfill({
    status: 200,
    contentType: 'application/vnd.apple.mpegurl',
    body: playlist,
  }));

  return { starts, stops };
}

async function loadHub(page: Page, did: string) {
  await page.locator(`.load-channels[data-did="${did}"]`).click();
  const dialog = page.locator('#code-dialog');
  await expect(dialog).toBeVisible();
  await page.locator('#security-code').fill(codes[did]);
  await page.locator('#code-submit').click();
  await expect(dialog).toBeHidden();
}

async function addHub(page: Page, did: string) {
  await page.locator('#manage-hubs').click();
  await page.locator('#add-hub').click();
  await expect(page.locator('#setup-section')).toBeVisible();
  await expect(page.locator('.hub-card')).toHaveCount(2);
  await loadHub(page, did);
}

async function loadBothHubs(page: Page) {
  await page.goto('/?lang=en');
  await page.getByRole('button', { name: 'Discover base stations' }).click();
  await expect(page.locator('.hub-card')).toHaveCount(2);
  await loadHub(page, 'CMAG-TEST-AAA');
  await addHub(page, 'CMAG-TEST-BBB');
}

test.describe('several base stations', () => {
  test.use({ colorScheme: 'light' });

  test('keeps every loaded base station in the wall', async ({ page }) => {
    // Arrange
    await mockGateway(page);

    // Act
    await loadBothHubs(page);

    // Assert
    await expect(page.locator('#wall-section')).toBeVisible();
    await expect(page.locator('#setup-section')).toBeHidden();
    await expect(page.locator('.camera-tile')).toHaveCount(3);
    await expect(page.locator('.camera-tile[data-hub-key="CMAG-TEST-BBB"]')).toHaveCount(1);
    await expect(page.locator('#hub-filters .filter-chip')).toHaveCount(3);
    await page.locator('#manage-hubs').click();
    await expect(page.locator('#hub-manager .hub-row')).toHaveCount(2);
  });

  test('starts streams on every base station at once', async ({ page }) => {
    // Arrange
    const { starts } = await mockGateway(page);
    await loadBothHubs(page);

    // Act
    await page.locator('#start-all-hubs').click();

    // Assert
    await expect(page.locator('.camera-tile .state-pill.live')).toHaveCount(3);
    expect(starts.map((item) => item.did).sort()).toEqual([
      'CMAG-TEST-AAA', 'CMAG-TEST-AAA', 'CMAG-TEST-BBB',
    ]);
    await expect(page.locator('#start-all-hubs')).toBeDisabled();
    await expect(page.locator('#stop-all-hubs')).toBeEnabled();
  });

  test('locks camera inventory while any stream is active', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await loadBothHubs(page);

    // Act
    await page.locator('#start-all-hubs').click();
    await expect(page.locator('.camera-tile .state-pill.live')).toHaveCount(3);

    // Assert
    await expect(page.locator('.load-channels:disabled')).toHaveCount(2);
    await expect(page.locator('#discover')).toBeDisabled();
    await page.locator('#manage-hubs').click();
    await expect(page.locator('#add-hub')).toBeDisabled();
  });

  test('stops streams on every base station at once', async ({ page }) => {
    // Arrange
    const { stops } = await mockGateway(page);
    await loadBothHubs(page);
    await page.locator('#start-all-hubs').click();
    await expect(page.locator('.camera-tile .state-pill.live')).toHaveCount(3);

    // Act
    await page.locator('#stop-all-hubs').click();

    // Assert
    await expect(page.locator('.camera-tile .state-pill.live')).toHaveCount(0);
    await expect(page.locator('.camera-tile .state-pill').first()).toContainText('Ready');
    expect(stops.sort()).toEqual([
      'CMAG-TEST-AAA-ch0', 'CMAG-TEST-AAA-ch1', 'CMAG-TEST-BBB-ch0',
    ]);
    await expect(page.locator('#discover')).toBeEnabled();
  });
});

test.describe('hub metadata', () => {
  test.use({ colorScheme: 'light' });

  test('shows complete DID and MAC values', async ({ page }) => {
    // Arrange
    const did = 'CMAG-018290-LYHFC-7B3D-9C2E';
    const mac = '8C:11:CB:5C:62:AF:19';
    await page.route('**/api/discover', (route) => route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        devices: [{ ip: '192.0.2.30', did, model: 'SMH-T1R' }],
        durationMs: 1,
      }),
    }));
    await page.route('**/api/channels', (route) => route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ hub: { mac }, channels: [{ index: 1, name: 'Camera' }] }),
    }));
    await page.goto('/?lang=en');
    await page.getByRole('button', { name: 'Discover base stations' }).click();
    await page.locator('.load-channels').click();
    await page.locator('#security-code').fill('code-a');
    await page.locator('#code-submit').click();
    await expect(page.locator('#wall-section')).toBeVisible();

    // Act: open the setup page again to inspect the loaded hub metadata.
    await page.locator('#manage-hubs').click();
    await page.locator('#add-hub').click();
    await expect(page.locator('.hub-card')).toHaveCount(1);

    // Assert: the details are visible without any interaction
    const subtitle = page.locator('.hub-card p');
    const macValue = page.locator('.hub-metadata div', { hasText: 'MAC' }).locator('strong');
    await expect(page.locator('.hub-metadata')).toBeVisible();
    await expect(subtitle).toContainText(did);
    await expect(macValue).toHaveText(mac);
    expect(await subtitle.evaluate((element) => element.scrollWidth <= element.clientWidth + 1)).toBe(true);
    expect(await macValue.evaluate((element) => element.scrollWidth <= element.clientWidth + 1)).toBe(true);
    // DID, IP and model stand in the header only; the grid must not repeat them
    await expect(page.locator('.hub-metadata')).not.toContainText('DID');
    await expect(page.locator('.hub-metadata')).not.toContainText('Model');
  });

  test('shows the MAC address in the base station manager', async ({ page }) => {
    // Arrange
    const mac = 'A4:2B:1C:9D:00:11';
    await page.route('**/api/discover', (route) => route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ devices: [{ ip: '192.0.2.30', did: 'CMAG-018290-LYHFC', model: 'SMH-T1R' }], durationMs: 1 }),
    }));
    await page.route('**/api/channels', (route) => route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ hub: { mac }, channels: [{ index: 1, name: 'Camera' }] }),
    }));
    await page.goto('/?lang=en');
    await page.getByRole('button', { name: 'Discover base stations' }).click();
    await page.locator('.load-channels').click();
    await page.locator('#security-code').fill('code-a');
    await page.locator('#code-submit').click();
    await expect(page.locator('#wall-section')).toBeVisible();

    // Act
    await page.locator('#manage-hubs').click();

    // Assert
    const row = page.locator('#hub-manager .hub-row[data-hub-key="CMAG-018290-LYHFC"]');
    await expect(row.locator('.hub-row-net').first()).toHaveText('IP 192.0.2.30');
    await expect(row.locator('.hub-row-net').last()).toContainText(mac);
  });
});