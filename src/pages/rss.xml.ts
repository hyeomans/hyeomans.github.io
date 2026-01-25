import type { APIRoute } from 'astro';
import { getCollection } from 'astro:content';
import { SITE_TITLE, SITE_DESCRIPTION, SITE_URL } from '../consts';

const escapeXml = (value: string) =>
	value
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&#39;');

export const GET: APIRoute = async () => {
	const posts = (await getCollection('posts'))
		.filter((post) => !post.data.draft)
		.sort((a, b) => b.data.pubDate.valueOf() - a.data.pubDate.valueOf());

	const items = posts
		.map((post) => {
			const slug = post.slug.split('/').pop() ?? post.slug;
			const link = post.data.lang === 'es'
				? `${SITE_URL}/es/posts/${slug}/`
				: `${SITE_URL}/posts/${slug}/`;
			const pubDate = post.data.pubDate.toUTCString();
			const description = escapeXml(post.data.description);
			const title = escapeXml(post.data.title);
			return `
				<item>
					<title>${title}</title>
					<link>${link}</link>
					<guid isPermaLink="true">${link}</guid>
					<pubDate>${pubDate}</pubDate>
					<description>${description}</description>
				</item>
			`.trim();
		})
		.join('');

	const lastBuildDate = posts[0]?.data.pubDate?.toUTCString() ?? new Date().toUTCString();

	const rss = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>${escapeXml(SITE_TITLE)}</title>
    <link>${SITE_URL}</link>
    <description>${escapeXml(SITE_DESCRIPTION)}</description>
    <language>en</language>
    <lastBuildDate>${lastBuildDate}</lastBuildDate>
    ${items}
  </channel>
</rss>`;

	return new Response(rss, {
		headers: {
			'Content-Type': 'application/xml; charset=utf-8',
		},
	});
};
