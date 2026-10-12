// Measures slow frames on a phone-sized screen and a desktop screen.
// Build the site the way the pages workflow does, serve it, then:
//   PLAYWRIGHT_CJS=/path/to/playwright/package.json node tools/lag/probe.mjs http://127.0.0.1:8765
// A frame over 32ms is slow. The report says which part of the picture it was.
import { createRequire } from 'node:module';

const require = createRequire(process.env.PLAYWRIGHT_CJS || import.meta.url);
const { chromium } = require('playwright');

const base = process.argv[2] || 'http://127.0.0.1:8765';
const floors = [1, 41, 73];

async function read(page) {
  return page.evaluate(() => window.crusaderLag.report());
}
async function reset(page) {
  await page.evaluate(() => window.crusaderLag.reset());
}
async function pan(page, key, ms) {
  await page.keyboard.down(key);
  await page.waitForTimeout(ms);
  await page.keyboard.up(key);
  await page.waitForTimeout(300);
}

async function floorRun(page, n) {
  await page.evaluate(() => window.crusaderLag.flat(true));
  const info = await page.evaluate((floor) => window.crusaderLag.floor(floor), n);
  await page.waitForTimeout(900);
  const open = await read(page);
  await reset(page);
  await pan(page, 'ArrowRight', 1200);
  const panFlat = await read(page);
  await reset(page);
  await page.evaluate(() => window.crusaderLag.zoom(1.8));
  await page.waitForTimeout(500);
  await pan(page, 'ArrowLeft', 1200);
  const zoomed = await read(page);
  await reset(page);
  await page.evaluate(() => window.crusaderLag.flat(false));
  await page.waitForTimeout(900);
  const isoOpen = await read(page);
  await reset(page);
  await page.evaluate(() => window.crusaderLag.zoom(1.8));
  await page.waitForTimeout(500);
  await pan(page, 'ArrowRight', 1200);
  const isoPan = await read(page);
  await reset(page);
  await page.evaluate(() => window.crusaderLag.whole());
  await page.waitForTimeout(900);
  const isoWhole = await read(page);
  return { info, open, panFlat, zoomed, isoOpen, isoPan, isoWhole };
}

const browser = await chromium.launch({ headless: true });
const screens = [
  ['phone', { width: 390, height: 844 }, 2],
  ['desktop', { width: 1280, height: 800 }, 1],
];
const out = {};
for (const [name, viewport, dsf] of screens) {
  const ctx = await browser.newContext({ viewport, deviceScaleFactor: dsf, hasTouch: name === 'phone' });
  const page = await ctx.newPage();
  page.on('pageerror', (e) => console.error(name, e.message));
  await page.goto(`${base}/play.html?lag=1`, { waitUntil: 'domcontentloaded', timeout: 60000 });
  await page.waitForFunction(() => window.crusaderLag, { timeout: 90000 });
  out[name] = {};
  for (const n of floors) out[name][n] = await floorRun(page, n);
  await ctx.close();
}
await browser.close();
console.log(JSON.stringify(out, null, 2));
