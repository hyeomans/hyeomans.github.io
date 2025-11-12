import { defineCollection, z } from 'astro:content';

const posts = defineCollection({
	type: 'content',
	schema: ({ image }) =>
		z.object({
			title: z.string(),
			description: z.string(),
			pubDate: z.coerce.date(),
			updatedDate: z.coerce.date().optional(),
			author: z.string().default('Hector Yeomans'),
			tags: z.array(z.string()).default([]),
			lang: z.enum(['en', 'es']),
			heroImage: image().optional(),
			heroAlt: z.string().optional(),
			draft: z.boolean().default(false),
			translationKey: z.string().optional(),
		}),
});

export const collections = { posts };
