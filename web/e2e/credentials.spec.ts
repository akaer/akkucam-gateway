import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';

// Fake credentials only. Real codes must never be committed to the repository.
const hubs = [
  { did: 'CMAG-TEST-AAA', ip: '192.0.2.10', model: 'SMH-T1R' },
  { did: 'CMAG-TEST-BBB', ip: '192.0.2.11', model: 'SMH-T1R' },
];

const codes: Record<string, string> = {
  'CMAG-TEST-AAA': 'code-a',
  'CMAG-TEST-BBB': 'code-b',
};

type Attempt = { did: string; code: string };

async function mockGateway(page: Page, options: { rejectFirst?: boolean } = {}) {
  const attempts: Attempt[] = [];
  let rejected = false;

  await page.route('**/api/discover', (route) => route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ devices: hubs, durationMs: 1 }),
  }));

  await page.route('**/api/channels', (route) => {
    const body = JSON.parse(route.request().postData() || '{}');
    attempts.push({ did: body.did, code: body.securityCode });

    if (options.rejectFirst && !rejected) {
      rejected = true;
      return route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'The security code was rejected.', code: 'unauthorized' }),
      });
    }

    if (codes[body.did] !== body.securityCode) {
      return route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'The security code was rejected.', code: 'unauthorized' }),
      });
    }

    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ hub: { did: body.did }, channels: [{ index: 0, name: 'Camera' }] }),
    });
  });

  return attempts;
}

async function discover(page: Page) {
  await page.goto('/?lang=en');
  await page.getByRole('button', { name: 'Discover base stations' }).click();
  await expect(page.locator('.hub-card')).toHaveCount(2);
}

async function loadHub(page: Page, did: string, code: string) {
  await page.locator(`.load-channels[data-did="${did}"]`).click();
  const dialog = page.locator('#code-dialog');
  await expect(dialog).toBeVisible();
  await expect(dialog.locator('.dialog-hub')).toContainText(did);
  await page.locator('#security-code').fill(code);
  await page.locator('#code-submit').click();
  await expect(dialog).toBeHidden();
}

async function openSetupForAdd(page: Page) {
  await page.locator('#manage-hubs').click();
  await page.locator('#add-hub').click();
  await expect(page.locator('#setup-section')).toBeVisible();
  await expect(page.locator('.hub-card')).toHaveCount(2);
}

async function waitForAttempts(attempts: Attempt[], count: number) {
  await expect.poll(() => attempts.length).toBe(count);
}

test.describe('security code visibility', () => {
  test.use({ colorScheme: 'light' });

  test('reveals the typed code and starts masked again', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await discover(page);
    await page.locator('.load-channels[data-did="CMAG-TEST-AAA"]').click();
    const dialog = page.locator('#code-dialog');
    const input = page.locator('#security-code');
    const toggle = page.locator('#code-visibility');
    await input.fill('code-a');

    // Assert: masked by default
    await expect(toggle).toHaveAttribute('aria-pressed', 'false');
    await expect(toggle).toHaveAttribute('aria-label', 'Show security code');
    await expect(input).toHaveAttribute('type', 'password');

    // Act and Assert: first click reveals, the value survives
    await toggle.click();
    await expect(toggle).toHaveAttribute('aria-pressed', 'true');
    await expect(input).toHaveAttribute('type', 'text');
    await expect(input).toHaveValue('code-a');

    // Act and Assert: second click masks again
    await toggle.click();
    await expect(input).toHaveAttribute('type', 'password');

    // Act: close and reopen the dialog
    await dialog.locator('[data-close]').first().click();
    await expect(dialog).toBeHidden();
    await page.locator('.load-channels[data-did="CMAG-TEST-AAA"]').click();

    // Assert: a fresh dialog never starts in the visible state
    await expect(dialog).toBeVisible();
    await expect(input).toHaveAttribute('type', 'password');
    await expect(toggle).toHaveAttribute('aria-pressed', 'false');
    await expect(input).toHaveValue('');
  });
});

test.describe('per-base-station security codes', () => {
  test.use({ colorScheme: 'light' });

  test('two base stations use separate security codes', async ({ page }) => {
    // Arrange
    const attempts = await mockGateway(page);
    await discover(page);

    // Act
    await loadHub(page, 'CMAG-TEST-AAA', 'code-a');
    await waitForAttempts(attempts, 1);
    await expect(page.locator('.camera-tile')).toHaveCount(1);

    await openSetupForAdd(page);
    await loadHub(page, 'CMAG-TEST-BBB', 'code-b');
    await waitForAttempts(attempts, 2);

    // Assert
    await expect(page.locator('.camera-tile[data-hub-key="CMAG-TEST-BBB"]')).toHaveCount(1);
    expect(attempts).toEqual([
      { did: 'CMAG-TEST-AAA', code: 'code-a' },
      { did: 'CMAG-TEST-BBB', code: 'code-b' },
    ]);
  });

  test('a base station reuses its stored code without prompting', async ({ page }) => {
    // Arrange
    const attempts = await mockGateway(page);
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA', 'code-a');
    await waitForAttempts(attempts, 1);

    // Act
    await openSetupForAdd(page);
    await page.locator('.load-channels[data-did="CMAG-TEST-AAA"]').click();
    await waitForAttempts(attempts, 2);

    // Assert
    await expect(page.locator('#code-dialog')).toBeHidden();
    expect(attempts[1]).toEqual({ did: 'CMAG-TEST-AAA', code: 'code-a' });
  });

  test('a rejected code is discarded and asked again', async ({ page }) => {
    // Arrange
    const attempts = await mockGateway(page, { rejectFirst: true });
    await discover(page);

    // Act
    await loadHub(page, 'CMAG-TEST-AAA', 'code-a');
    await waitForAttempts(attempts, 1);

    // Assert
    await expect(page.locator('#notice')).toHaveClass(/error/);
    await expect(page.locator('.hub-credential[data-hub-key="CMAG-TEST-AAA"]')).toBeHidden();

    // Act
    await loadHub(page, 'CMAG-TEST-AAA', 'code-a');
    await waitForAttempts(attempts, 2);

    // Assert
    await expect(page.locator('.camera-tile')).toHaveCount(1);
    expect(attempts).toEqual([
      { did: 'CMAG-TEST-AAA', code: 'code-a' },
      { did: 'CMAG-TEST-AAA', code: 'code-a' },
    ]);
  });

  test('forgetting a base station keeps the other code', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA', 'code-a');
    await openSetupForAdd(page);
    await loadHub(page, 'CMAG-TEST-BBB', 'code-b');

    await page.locator('#manage-hubs').click();
    const rowA = page.locator('#hub-manager .hub-row[data-hub-key="CMAG-TEST-AAA"]');
    const rowB = page.locator('#hub-manager .hub-row[data-hub-key="CMAG-TEST-BBB"]');
    await expect(rowA.locator('.code-state')).toHaveClass(/active/);
    await expect(rowB.locator('.code-state')).toHaveClass(/active/);

    // Act
    await rowA.getByRole('button', { name: 'Remove code' }).click();

    // Assert
    await expect(rowA.locator('.code-state')).not.toHaveClass(/active/);
    await expect(rowB.locator('.code-state')).toHaveClass(/active/);
    await page.locator('#add-hub').click();
    await expect(page.locator('#setup-section')).toBeVisible();
    await page.locator('.load-channels[data-did="CMAG-TEST-AAA"]').click();
    await expect(page.locator('#code-dialog')).toBeVisible();
  });

  test('forgets all codes from the hub manager', async ({ page }) => {
    // Arrange
    await mockGateway(page);
    await discover(page);
    await loadHub(page, 'CMAG-TEST-AAA', 'code-a');
    await openSetupForAdd(page);
    await loadHub(page, 'CMAG-TEST-BBB', 'code-b');
    await page.locator('#manage-hubs').click();
    const rowA = page.locator('#hub-manager .hub-row[data-hub-key="CMAG-TEST-AAA"]');
    const rowB = page.locator('#hub-manager .hub-row[data-hub-key="CMAG-TEST-BBB"]');
    await expect(rowA.locator('.code-state')).toHaveClass(/active/);
    await expect(rowB.locator('.code-state')).toHaveClass(/active/);

    // Act: the first click only arms the destructive action (ADR-0028).
    await page.locator('#forget-all-codes').click();

    // Assert
    await expect(page.locator('#forget-all-codes')).toHaveText('Remove all codes now?');
    await expect(rowA.locator('.code-state')).toHaveClass(/active/);
    await page.locator('#forget-all-codes').click();
    await expect(rowA.locator('.code-state')).not.toHaveClass(/active/);
    await expect(rowB.locator('.code-state')).not.toHaveClass(/active/);
    await page.locator('#add-hub').click();
    await expect(page.locator('#setup-section')).toBeVisible();
    await page.locator('.load-channels[data-did="CMAG-TEST-AAA"]').click();
    await expect(page.locator('#code-dialog')).toBeVisible();
  });
});