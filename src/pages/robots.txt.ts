import type { APIRoute } from 'astro';
import { SITE_URL } from '../consts';

export const GET: APIRoute = () => {
	const robotsTxt = `
# Allow all robots
User-agent: *
Allow: /

# Disallow Cloudflare internal paths
Disallow: /cdn-cgi/

# Sitemap
Sitemap: ${SITE_URL}/sitemap-index.xml
`.trim();

	return new Response(robotsTxt, {
		headers: {
			'Content-Type': 'text/plain; charset=utf-8',
		},
	});
};
