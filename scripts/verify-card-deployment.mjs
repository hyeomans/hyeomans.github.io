const baseUrl = process.argv[2];
if (!baseUrl) throw new Error('Usage: node scripts/verify-card-deployment.mjs <pages-url>');

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
for (const required of ['Hector Yeomans', 'me@hyeomans.com', 'linkedin.com/in/hector-yeomans']) {
  if (!cardHtml.includes(required)) throw new Error(`Deployed card is missing ${required}`);
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
