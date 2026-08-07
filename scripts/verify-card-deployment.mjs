const baseUrl = process.argv[2];
if (!baseUrl) throw new Error('Usage: node scripts/verify-card-deployment.mjs <pages-url>');

const expectedEmail = 'me@hyeomans.com';

const decodeCloudflareEmail = (encoded) => {
  const key = Number.parseInt(encoded.slice(0, 2), 16);
  let decoded = '';
  for (let index = 2; index < encoded.length; index += 2) {
    decoded += String.fromCharCode(Number.parseInt(encoded.slice(index, index + 2), 16) ^ key);
  }
  return decoded;
};

const fetchWithRetry = async (url, attempts = 5) => {
  let response;
  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    response = await fetch(url, { redirect: 'follow' });
    if (response.ok) return response;
    if (attempt < attempts) await new Promise((resolve) => setTimeout(resolve, 2000));
  }
  return response;
};

const cardUrl = new URL('card/', baseUrl);
const cardResponse = await fetchWithRetry(cardUrl);
if (!cardResponse.ok) throw new Error(`Card deployment returned HTTP ${cardResponse.status}`);

const cardHtml = await cardResponse.text();
for (const required of ['Hector Yeomans', 'linkedin.com/in/hector-yeomans']) {
  if (!cardHtml.includes(required)) throw new Error(`Deployed card is missing ${required}`);
}

const encodedEmails = [...cardHtml.matchAll(/data-cfemail="([a-f0-9]+)"/gi)].map((match) => decodeCloudflareEmail(match[1]));
if (!cardHtml.includes(expectedEmail) && !encodedEmails.includes(expectedEmail)) {
  throw new Error(`Deployed card is missing ${expectedEmail}`);
}
if (/tel:|type=["']tel["']|itemprop=["']telephone["']|\d{3}[ .-]\d{3}[ .-]\d{4}/i.test(cardHtml)) {
  throw new Error('Deployed card contains telephone data');
}

const applePassUrl = new URL('wallet/hector-yeomans.pkpass', baseUrl);
const applePassResponse = await fetch(applePassUrl, { method: 'HEAD', redirect: 'follow' });

if (applePassResponse.status === 200) {
  const contentType = applePassResponse.headers.get('content-type')?.split(';')[0].trim();
  if (contentType !== 'application/vnd.apple.pkpass') {
    throw new Error(`Apple pass has incorrect Content-Type: ${contentType || '(missing)'}`);
  }
  console.log('Deployed Apple Wallet MIME type verified');
} else if (applePassResponse.status === 404) {
  console.log('Apple Wallet pass is not configured; MIME check skipped');
} else {
  throw new Error(`Apple pass endpoint returned HTTP ${applePassResponse.status}`);
}

console.log(`Deployed contact card verified at ${cardResponse.url}`);
