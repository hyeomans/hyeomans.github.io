import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs';
import { relative, join } from 'node:path';

const SITE = 'https://hyeomans.com';
const DIST = 'dist';

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

const routeForFile = (file) => {
	const rel = relative(DIST, file);
	if (rel.endsWith('index.html')) {
		const route = `/${rel.slice(0, -'index.html'.length)}`;
		return route === '/' ? route : route.replace(/\/?$/, '/');
	}
	return `/${rel}`;
};

const attr = (tag, name) => {
	const match = tag.match(new RegExp(`\\s${name}=(["'])(.*?)\\1`, 'i'));
	return match?.[2];
};

const text = (html, pattern) => html.match(pattern)?.[1]?.replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim() ?? '';

const failures = [];
const fail = (message) => failures.push(message);

if (!existsSync(DIST)) {
	fail('dist/ does not exist. Run npm run build first.');
} else {
	const pages = walk(DIST, (path) => path.endsWith('.html')).map((file) => {
		const html = readFileSync(file, 'utf8');
		const route = routeForFile(file);
		const title = text(html, /<title[^>]*>([\s\S]*?)<\/title>/i);
		const description = html.match(/<meta\s+[^>]*name=["']description["'][^>]*>/i)?.[0] ?? '';
		const robots = attr(html.match(/<meta\s+[^>]*name=["']robots["'][^>]*>/i)?.[0] ?? '', 'content') ?? '';
		const canonical = [...html.matchAll(/<link\s+[^>]*rel=["']canonical["'][^>]*>/gi)].map((match) => attr(match[0], 'href'));
		const h1 = [...html.matchAll(/<h1\b[\s\S]*?<\/h1>/gi)].map((match) => match[0]);
		const imgs = [...html.matchAll(/<img\b[^>]*>/gi)].map((match) => match[0]);
		const anchors = [...html.matchAll(/<a\b[^>]*>/gi)].map((match) => match[0]);
		return { file, route, html, title, description: attr(description, 'content') ?? '', robots, canonical, h1, imgs, anchors };
	});

	const indexable = pages.filter((page) => !page.robots.toLowerCase().includes('noindex') && page.route !== '/404.html');

	for (const page of indexable) {
		if (page.title.length < 20 || page.title.length > 75) {
			fail(`${page.route} has title length ${page.title.length}: ${page.title}`);
		}
		if (page.description.length < 50 || page.description.length > 180) {
			fail(`${page.route} has description length ${page.description.length}: ${page.description}`);
		}
		if (page.h1.length !== 1) {
			fail(`${page.route} has ${page.h1.length} H1 elements`);
		}
		if (page.canonical.length !== 1 || !page.canonical[0]?.startsWith(SITE)) {
			fail(`${page.route} has invalid canonical: ${page.canonical.join(', ')}`);
		}
		for (const img of page.imgs) {
			if (attr(img, 'alt') === undefined) {
				fail(`${page.route} has an image without alt: ${img}`);
			}
		}
	}

	const existing = new Set(['/robots.txt', '/rss.xml', '/sitemap-index.xml']);
	for (const file of walk(DIST)) {
		const rel = `/${relative(DIST, file)}`;
		existing.add(rel);
		if (rel.endsWith('/index.html')) existing.add(rel.slice(0, -'index.html'.length));
	}

	for (const page of pages) {
		for (const anchor of page.anchors) {
			const href = attr(anchor, 'href');
			if (!href || href.startsWith('#') || /^(mailto|tel|javascript):/i.test(href)) continue;

			const url = new URL(href, `${SITE}${page.route}`);
			if (url.hostname !== 'hyeomans.com') continue;

			const path = url.pathname;
			const candidates = new Set([path, path.replace(/\/?$/, '/'), `${path.replace(/\/$/, '')}/index.html`]);
			if (![...candidates].some((candidate) => existing.has(candidate))) {
				fail(`${page.route} links to missing internal URL: ${href}`);
			}
		}
	}

	const noindexRoutes = new Set(pages.filter((page) => page.robots.toLowerCase().includes('noindex')).map((page) => `${SITE}${page.route}`));
	for (const sitemap of walk(DIST, (path) => /sitemap.*\.xml$/.test(path))) {
		const xml = readFileSync(sitemap, 'utf8');
		for (const loc of xml.matchAll(/<loc>(.*?)<\/loc>/g)) {
			if (noindexRoutes.has(loc[1])) {
				fail(`${relative(DIST, sitemap)} includes noindex URL: ${loc[1]}`);
			}
		}
	}
}

if (failures.length) {
	console.error(`SEO audit failed with ${failures.length} issue(s):`);
	for (const failure of failures) console.error(`- ${failure}`);
	process.exit(1);
}

console.log('SEO audit passed.');
