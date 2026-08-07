import { createHash, createSign, createVerify } from 'node:crypto';
import { mkdir, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';
import { spawn } from 'node:child_process';

const root = resolve(import.meta.dirname, '..');
const publicWalletDirectory = resolve(root, 'public/wallet');
const linksFile = resolve(root, 'src/data/generated-wallet-links.json');
const requireAll = process.argv.includes('--require-all');

const CARD = Object.freeze({
  name: 'Hector Yeomans',
  role: 'Staff Software Engineer',
  email: 'me@hyeomans.com',
  linkedin: 'https://www.linkedin.com/in/hector-yeomans/',
  url: 'https://hyeomans.com/card/',
});

const decodeSecret = (name) => {
  const value = process.env[name]?.trim();
  if (!value) return null;
  return Buffer.from(value.includes('-----BEGIN') ? value : Buffer.from(value, 'base64').toString('utf8'));
};

const base64url = (value) => Buffer.from(value).toString('base64url');

const signGoogleJwt = (credentials, payload) => {
  const header = base64url(JSON.stringify({ alg: 'RS256', typ: 'JWT', kid: credentials.private_key_id }));
  const body = base64url(JSON.stringify(payload));
  const unsignedToken = `${header}.${body}`;
  const signature = createSign('RSA-SHA256').update(unsignedToken).end().sign(credentials.private_key);
  return `${unsignedToken}.${signature.toString('base64url')}`;
};

const localized = (value) => ({
  defaultValue: { language: 'en-US', value },
});

const run = (command, args, options = {}) => new Promise((resolvePromise, reject) => {
  const child = spawn(command, args, { ...options, stdio: ['ignore', 'pipe', 'pipe'] });
  let stderr = '';
  child.stderr.on('data', (chunk) => { stderr += chunk; });
  child.on('error', reject);
  child.on('close', (code) => {
    if (code === 0) resolvePromise();
    else reject(new Error(`${command} exited with ${code}${stderr ? `: ${stderr.trim()}` : ''}`));
  });
});

const createApplePass = async () => {
  const passTypeIdentifier = process.env.APPLE_PASS_TYPE_IDENTIFIER?.trim();
  const teamIdentifier = process.env.APPLE_TEAM_IDENTIFIER?.trim();
  const wwdr = decodeSecret('APPLE_WWDR_CERTIFICATE');
  const signerCert = decodeSecret('APPLE_SIGNER_CERTIFICATE');
  const signerKey = decodeSecret('APPLE_SIGNER_KEY');

  if (!passTypeIdentifier || !teamIdentifier || !wwdr || !signerCert || !signerKey) return null;

  const passJson = {
    formatVersion: 1,
    passTypeIdentifier,
    serialNumber: 'hector-yeomans-contact-v1',
    teamIdentifier,
    organizationName: CARD.name,
    description: `${CARD.name} contact card`,
    logoText: CARD.name,
    backgroundColor: 'rgb(21, 21, 21)',
    foregroundColor: 'rgb(246, 240, 229)',
    labelColor: 'rgb(224, 157, 31)',
    sharingProhibited: false,
    generic: {
      primaryFields: [{ key: 'name', label: 'CONTACT', value: CARD.name }],
      secondaryFields: [{ key: 'role', label: 'ROLE', value: CARD.role }],
      auxiliaryFields: [{ key: 'email', label: 'EMAIL', value: CARD.email }],
      backFields: [
        { key: 'email-link', label: 'Email', value: CARD.email },
        { key: 'linkedin', label: 'LinkedIn', value: CARD.linkedin },
        { key: 'website', label: 'Contact card', value: CARD.url },
        { key: 'privacy', label: 'Privacy', value: 'This card contains no phone number or location data.' },
      ],
    },
    barcodes: [{
      format: 'PKBarcodeFormatQR',
      message: CARD.url,
      messageEncoding: 'iso-8859-1',
      altText: 'hyeomans.com/card',
    }],
  };

  const output = resolve(publicWalletDirectory, 'hector-yeomans.pkpass');
  const temporaryDirectory = await mkdtemp(resolve(tmpdir(), 'hyeomans-pass-'));

  try {
    const files = {
      'pass.json': Buffer.from(JSON.stringify(passJson)),
      'icon.png': await readFile(resolve(root, 'public/card/apple-icon.png')),
      'icon@2x.png': await readFile(resolve(root, 'public/card/apple-icon@2x.png')),
      'logo.png': await readFile(resolve(root, 'public/card/apple-logo.png')),
      'logo@2x.png': await readFile(resolve(root, 'public/card/apple-logo@2x.png')),
    };

    await Promise.all(Object.entries(files).map(([name, contents]) => writeFile(resolve(temporaryDirectory, name), contents)));

    const manifest = Object.fromEntries(
      Object.entries(files).map(([name, contents]) => [name, createHash('sha1').update(contents).digest('hex')]),
    );
    const manifestPath = resolve(temporaryDirectory, 'manifest.json');
    const wwdrPath = resolve(temporaryDirectory, 'wwdr.pem');
    const certificatePath = resolve(temporaryDirectory, 'signer-certificate.pem');
    const keyPath = resolve(temporaryDirectory, 'signer-key.pem');
    const signaturePath = resolve(temporaryDirectory, 'signature');

    await Promise.all([
      writeFile(manifestPath, JSON.stringify(manifest)),
      writeFile(wwdrPath, wwdr),
      writeFile(certificatePath, signerCert),
      writeFile(keyPath, signerKey, { mode: 0o600 }),
    ]);

    const signArguments = [
      'smime', '-binary', '-sign',
      '-certfile', wwdrPath,
      '-signer', certificatePath,
      '-inkey', keyPath,
      '-in', manifestPath,
      '-out', signaturePath,
      '-outform', 'DER',
    ];
    if (process.env.APPLE_SIGNER_KEY_PASSPHRASE) {
      signArguments.push('-passin', 'env:APPLE_SIGNER_KEY_PASSPHRASE');
    }

    await run('openssl', signArguments);
    await run('openssl', [
      'smime', '-verify', '-inform', 'DER',
      '-in', signaturePath,
      '-content', manifestPath,
      '-noverify',
      '-out', resolve(temporaryDirectory, 'verified-manifest.json'),
    ]);

    await rm(resolve(temporaryDirectory, 'verified-manifest.json'));
    await rm(wwdrPath);
    await rm(certificatePath);
    await rm(keyPath);
    await rm(output, { force: true });
    await run('zip', ['-X', '-q', '-r', output, '.'], { cwd: temporaryDirectory });
    return '/wallet/hector-yeomans.pkpass';
  } finally {
    await rm(temporaryDirectory, { recursive: true, force: true });
  }
};

const createGooglePass = async () => {
  const issuerId = process.env.GOOGLE_WALLET_ISSUER_ID?.trim();
  const credentialsBuffer = decodeSecret('GOOGLE_WALLET_CREDENTIALS');
  if (!issuerId || !credentialsBuffer) return null;

  const credentials = JSON.parse(credentialsBuffer.toString('utf8'));
  if (!credentials.client_email || !credentials.private_key) {
    throw new Error('GOOGLE_WALLET_CREDENTIALS is missing client_email or private_key');
  }

  const classId = `${issuerId}.hector_yeomans_contact`;
  const objectId = `${issuerId}.hector_yeomans_contact_v1`;
  const genericClass = { id: classId };
  const genericObject = {
    id: objectId,
    classId,
    state: 'ACTIVE',
    cardTitle: localized(CARD.name),
    header: localized(CARD.role),
    hexBackgroundColor: '#151515',
    barcode: {
      type: 'QR_CODE',
      value: CARD.url,
      alternateText: 'hyeomans.com/card',
    },
    linksModuleData: {
      uris: [
        { id: 'email', uri: `mailto:${CARD.email}`, description: CARD.email },
        { id: 'linkedin', uri: CARD.linkedin, description: 'LinkedIn' },
      ],
    },
  };

  const token = signGoogleJwt(credentials, {
    iss: credentials.client_email,
    aud: 'google',
    typ: 'savetowallet',
    iat: Math.floor(Date.now() / 1000),
    origins: ['https://hyeomans.com'],
    payload: { genericClasses: [genericClass], genericObjects: [genericObject] },
  });
  const saveUrl = `https://pay.google.com/gp/v/save/${token}`;

  if (saveUrl.length > 1800) {
    throw new Error(`Google Wallet save URL is ${saveUrl.length} characters; the supported maximum is 1800`);
  }

  const [encodedHeader, encodedPayload, encodedSignature] = token.split('.');
  const verifier = createVerify('RSA-SHA256').update(`${encodedHeader}.${encodedPayload}`).end();
  if (!verifier.verify(credentials.private_key, Buffer.from(encodedSignature, 'base64url'))) {
    throw new Error('Google Wallet JWT signature failed local verification');
  }

  return saveUrl;
};

await mkdir(publicWalletDirectory, { recursive: true });
await rm(resolve(publicWalletDirectory, 'hector-yeomans.pkpass'), { force: true });

const results = { apple: null, google: null, generatedAt: null };
const failures = [];

for (const [platform, generate] of [['apple', createApplePass], ['google', createGooglePass]]) {
  try {
    results[platform] = await generate();
  } catch (error) {
    failures.push(`${platform}: ${error instanceof Error ? error.message : String(error)}`);
  }
}

if (results.apple || results.google) results.generatedAt = new Date().toISOString();
await writeFile(linksFile, `${JSON.stringify(results, null, 2)}\n`);

if (failures.length) throw new Error(`Wallet generation failed:\n${failures.join('\n')}`);
if (requireAll && (!results.apple || !results.google)) {
  throw new Error('Both Apple and Google issuer credentials are required for wallet:verify');
}

console.log(`Apple Wallet: ${results.apple ? 'generated' : 'not configured'}`);
console.log(`Google Wallet: ${results.google ? 'generated' : 'not configured'}`);
