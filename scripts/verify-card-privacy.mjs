import { readFile } from 'node:fs/promises';
import { resolve } from 'node:path';

const root = resolve(import.meta.dirname, '..');
const cardHtml = await readFile(resolve(root, 'dist/card/index.html'), 'utf8');
const links = JSON.parse(await readFile(resolve(root, 'src/data/generated-wallet-links.json'), 'utf8'));

const required = [
  'Hector Yeomans',
  'Staff Software Engineer',
  'me@hyeomans.com',
  'https://www.linkedin.com/in/hector-yeomans/',
  '/card/qr.svg',
  'noindex,nofollow,noarchive,nosnippet',
];

const forbidden = [
  ['telephone URI', /tel:/i],
  ['telephone input or schema field', /(?:type=["']tel["']|itemprop=["']telephone["']|"telephone"\s*:)/i],
  ['North American telephone number', /(?:\+?1[ .-]?)?\(?\d{3}\)?[ .-]\d{3}[ .-]\d{4}/],
  ['unapproved GitHub profile', /github\.com\/hyeomans/i],
  ['unapproved Twitter profile', /(?:twitter\.com|x\.com)\/h_yeomans/i],
  ['Google Analytics', /googletagmanager|google-analytics|G-CQCQ4X7JHE/i],
];

for (const value of required) {
  if (!cardHtml.includes(value)) throw new Error(`Card is missing required value: ${value}`);
}

for (const [label, pattern] of forbidden) {
  if (pattern.test(cardHtml)) throw new Error(`Card contains forbidden ${label}`);
}

if (links.apple && !cardHtml.includes(links.apple)) throw new Error('Apple Wallet link was not rendered');
if (links.google && !cardHtml.includes(links.google.replaceAll('&', '&amp;'))) {
  throw new Error('Google Wallet link was not rendered');
}

console.log('Card privacy allowlist verified');
