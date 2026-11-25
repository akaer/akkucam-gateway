import { defineConfig, devices } from '@playwright/test';

const baseURL = process.env.AKKUCAM_BASE_URL || 'http://127.0.0.1:8080';
const chromiumExecutable = process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE;
const captureArtifacts = process.env.PLAYWRIGHT_NO_ARTIFACTS !== '1';

type Suite = {
  name: string;
  testMatch: RegExp;
  timeout?: number;
  use?: { trace?: 'off'; video?: 'off'; screenshot?: 'off' };
};

// Each suite runs in Chromium and Firefox. The Chromium projects keep their
// historical short names (`i18n`, `wall`, ...) so existing commands keep
// working; the Firefox projects add the `-firefox` suffix.
const suites: Suite[] = [
  { name: 'i18n', testMatch: /i18n\.spec\.ts/ },
  { name: 'branding', testMatch: /branding\.spec\.ts/ },
  { name: 'credentials', testMatch: /credentials\.spec\.ts/ },
  { name: 'workspaces', testMatch: /workspaces\.spec\.ts/ },
  { name: 'wall', testMatch: /wall\.spec\.ts/ },
  {
    name: 'live',
    testMatch: /live\.spec\.ts/,
    timeout: 120_000,
    // The live suite receives the security code. Never write traces, videos
    // or screenshots that could contain it.
    use: { trace: 'off', video: 'off', screenshot: 'off' },
  },
];

const browsers = [
  { suffix: '', device: 'Desktop Chrome', chromium: true },
  { suffix: '-firefox', device: 'Desktop Firefox', chromium: false },
] as const;

const projects = browsers.flatMap((browser) => suites.map((suite) => ({
  name: `${suite.name}${browser.suffix}`,
  testMatch: suite.testMatch,
  ...(suite.timeout ? { timeout: suite.timeout } : {}),
  use: {
    ...devices[browser.device],
    ...(browser.chromium && chromiumExecutable
      ? { launchOptions: { executablePath: chromiumExecutable } }
      : {}),
    ...suite.use,
  },
})));

export default defineConfig({
  testDir: '.',
  outputDir: 'test-results',
  timeout: 30_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI
    ? [['list'], ['html', { open: 'never' }]]
    : [['list']],
  use: {
    baseURL,
    trace: captureArtifacts ? 'retain-on-failure' : 'off',
    screenshot: captureArtifacts ? 'only-on-failure' : 'off',
    video: captureArtifacts ? 'retain-on-failure' : 'off',
  },
  projects,
});
