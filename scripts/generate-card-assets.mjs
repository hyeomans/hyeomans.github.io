import { mkdir, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import QRCode from 'qrcode';
import sharp from 'sharp';

const root = resolve(import.meta.dirname, '..');
const outputDirectory = resolve(root, 'public/card');
const cardUrl = 'https://hyeomans.com/card/';

const mark = (size) => `
<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 256 256">
  <rect width="256" height="256" rx="52" fill="#151515"/>
  <path d="M60 54h28v58h80V54h28v148h-28v-64H88v64H60z" fill="#f6f0e5"/>
  <circle cx="196" cy="54" r="14" fill="#e09d1f"/>
</svg>`;

await mkdir(outputDirectory, { recursive: true });

const qrSvg = await QRCode.toString(cardUrl, {
  type: 'svg',
  errorCorrectionLevel: 'H',
  margin: 1,
  color: { dark: '#151515', light: '#F6F0E5' },
});

await Promise.all([
  writeFile(resolve(outputDirectory, 'qr.svg'), qrSvg),
  writeFile(resolve(outputDirectory, 'mark.svg'), mark(256)),
  sharp(Buffer.from(mark(58))).resize(29, 29).png().toFile(resolve(outputDirectory, 'apple-icon.png')),
  sharp(Buffer.from(mark(116))).resize(58, 58).png().toFile(resolve(outputDirectory, 'apple-icon@2x.png')),
  sharp(Buffer.from(mark(160))).resize(160, 50, { fit: 'contain' }).png().toFile(resolve(outputDirectory, 'apple-logo.png')),
  sharp(Buffer.from(mark(320))).resize(320, 100, { fit: 'contain' }).png().toFile(resolve(outputDirectory, 'apple-logo@2x.png')),
]);

console.log(`Generated card assets for ${cardUrl}`);
