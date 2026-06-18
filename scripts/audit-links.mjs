import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { request } from 'node:https';
import { request as httpRequest } from 'node:http';

const DIST = 'dist';
const SITE_HOST = 'hyeomans.com';
const ALLOWED_BLOCKED_HOSTS = new Set([
	'onedrive.live.com',
	'www.npmjs.com',
]);

const walk = (dir, predicate = () => true) => {
	const files = [];
	for (const entry of readdirSync(dir)) {
		const path = join(dir, entry);
		const stats = statSync(path);
		if (stats.isDirectory()) {
			files.push(...walk(path, predicate));
		} else if (predicate(path)) {
			files.push(path);
		}
	}
	return files;
};

const attr = (tag, name) => {
	const match = tag.match(new RegExp(`\\s${name}=(["'])(.*?)\\1`, 'i'));
	return match?.[2]
		?.replaceAll('&amp;', '&')
		.replaceAll('&#38;', '&')
		.replaceAll('&#x26;', '&');
};

const fetchStatus = (url, method = 'HEAD', redirectCount = 0) =>
	new Promise((resolve) => {
		const parsed = new URL(url);
		const client = parsed.protocol === 'http:' ? httpRequest : request;
		const req = client(
			parsed,
			{
				method,
				headers: {
					'user-agent': 'Mozilla/5.0 hyeomans-seo-audit/1.0',
					accept: 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
				},
				timeout: 12000,
			},
			(res) => {
				res.resume();
				const location = res.headers.location;
				if (location && [301, 302, 303, 307, 308].includes(res.statusCode ?? 0) && redirectCount < 5) {
					resolve(fetchStatus(new URL(location, parsed).toString(), method, redirectCount + 1));
					return;
				}
				resolve({ status: res.statusCode ?? 0, finalUrl: parsed.toString() });
			},
		);

		req.on('timeout', () => {
			req.destroy();
			resolve({ status: 0, finalUrl: parsed.toString() });
		});
		req.on('error', () => resolve({ status: 0, finalUrl: parsed.toString() }));
		req.end();
	});

const externalUrls = new Set();
for (const file of walk(DIST, (path) => path.endsWith('.html'))) {
	const html = readFileSync(file, 'utf8');
	for (const match of html.matchAll(/<a\b[^>]*>/gi)) {
		const href = attr(match[0], 'href');
		if (!href || !/^https?:\/\//i.test(href)) continue;
		const url = new URL(href);
		if (url.hostname === SITE_HOST || url.hostname === `www.${SITE_HOST}`) continue;
		externalUrls.add(url.toString());
	}
}

const failures = [];
for (const url of [...externalUrls].sort()) {
	let result = await fetchStatus(url);
	if ([0, 403, 405].includes(result.status)) {
		result = await fetchStatus(url, 'GET');
	}

	const host = new URL(url).hostname;
	if (result.status === 403 && ALLOWED_BLOCKED_HOSTS.has(host)) continue;

	if (result.status === 0 || result.status >= 400) {
		failures.push(`${result.status} ${url}`);
	}
}

if (failures.length) {
	console.error(`External link audit failed with ${failures.length} issue(s):`);
	for (const failure of failures) console.error(`- ${failure}`);
	process.exit(1);
}

console.log(`External link audit passed for ${externalUrls.size} unique external URL(s).`);
