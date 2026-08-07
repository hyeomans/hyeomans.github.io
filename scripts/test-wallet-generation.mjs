import { mkdtemp, readFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';
import { spawn } from 'node:child_process';

const root = resolve(import.meta.dirname, '..');
const temporaryDirectory = await mkdtemp(resolve(tmpdir(), 'hyeomans-wallet-test-'));

const run = (command, args, options = {}) => new Promise((resolvePromise, reject) => {
  const child = spawn(command, args, { ...options, stdio: ['ignore', 'pipe', 'pipe'] });
  let stdout = '';
  let stderr = '';
  child.stdout.on('data', (chunk) => { stdout += chunk; });
  child.stderr.on('data', (chunk) => { stderr += chunk; });
  child.on('error', reject);
  child.on('close', (code) => {
    if (code === 0) resolvePromise(stdout);
    else reject(new Error(`${command} exited with ${code}: ${stderr.trim()}`));
  });
});

const path = (name) => resolve(temporaryDirectory, name);

try {
  await run('openssl', [
    'req', '-x509', '-newkey', 'rsa:2048', '-nodes',
    '-keyout', path('wwdr-key.pem'), '-out', path('wwdr.pem'),
    '-days', '1', '-subj', '/CN=Disposable Wallet Test CA',
  ]);
  await run('openssl', [
    'req', '-newkey', 'rsa:2048', '-nodes',
    '-keyout', path('signer-key.pem'), '-out', path('signer.csr'),
    '-subj', '/CN=Hector Yeomans/OU=XPHAU2VZR8/UID=pass.com.hyeomans.contact',
  ]);
  await run('openssl', [
    'x509', '-req', '-in', path('signer.csr'),
    '-CA', path('wwdr.pem'), '-CAkey', path('wwdr-key.pem'), '-CAcreateserial',
    '-out', path('signer.pem'), '-days', '1', '-sha256',
  ]);

  const [wwdr, signerCertificate, signerKey] = await Promise.all([
    readFile(path('wwdr.pem'), 'utf8'),
    readFile(path('signer.pem'), 'utf8'),
    readFile(path('signer-key.pem'), 'utf8'),
  ]);
  const googleCredentials = JSON.stringify({
    type: 'service_account',
    private_key_id: 'disposable-test-key',
    private_key: signerKey,
    client_email: 'wallet-test@hyeomans.invalid',
  });

  await run(process.execPath, ['scripts/generate-wallet-passes.mjs', '--require-all'], {
    cwd: root,
    env: {
      ...process.env,
      APPLE_PASS_TYPE_IDENTIFIER: 'pass.com.hyeomans.contact',
      APPLE_TEAM_IDENTIFIER: 'XPHAU2VZR8',
      APPLE_WWDR_CERTIFICATE: wwdr,
      APPLE_SIGNER_CERTIFICATE: signerCertificate,
      APPLE_SIGNER_KEY: signerKey,
      GOOGLE_WALLET_ISSUER_ID: '1234567890123456789',
      GOOGLE_WALLET_CREDENTIALS: googleCredentials,
    },
  });

  const links = JSON.parse(await readFile(resolve(root, 'src/data/generated-wallet-links.json'), 'utf8'));
  if (links.apple !== '/wallet/hector-yeomans.pkpass') throw new Error('Apple pass URL was not generated');
  if (!links.google?.startsWith('https://pay.google.com/gp/v/save/')) throw new Error('Google save URL was not generated');

  const passJson = JSON.parse(await run('unzip', ['-p', resolve(root, 'public/wallet/hector-yeomans.pkpass'), 'pass.json']));
  const publicPayload = JSON.stringify(passJson);
  for (const required of ['Hector Yeomans', 'me@hyeomans.com', 'linkedin.com/in/hector-yeomans']) {
    if (!publicPayload.includes(required)) throw new Error(`Apple pass is missing ${required}`);
  }
  if (/tel:|"telephone"\s*:|\d{3}[ .-]\d{3}[ .-]\d{4}/i.test(publicPayload)) {
    throw new Error('Apple pass contains telephone data');
  }

  const googleToken = new URL(links.google).pathname.split('/').pop();
  const googlePayload = JSON.parse(Buffer.from(googleToken.split('.')[1], 'base64url').toString('utf8'));
  const serializedGooglePayload = JSON.stringify(googlePayload);
  for (const required of ['Hector Yeomans', 'me@hyeomans.com', 'linkedin.com/in/hector-yeomans']) {
    if (!serializedGooglePayload.includes(required)) throw new Error(`Google pass is missing ${required}`);
  }
  if (/tel:|"telephone"\s*:|\d{3}[ .-]\d{3}[ .-]\d{4}/i.test(serializedGooglePayload)) {
    throw new Error('Google pass contains telephone data');
  }

  console.log('Apple and Google wallet generation verified with disposable credentials');
} finally {
  await run(process.execPath, ['scripts/generate-wallet-passes.mjs'], { cwd: root, env: process.env }).catch(() => {});
  await rm(temporaryDirectory, { recursive: true, force: true });
}
