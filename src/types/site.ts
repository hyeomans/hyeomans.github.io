export type SupportedLang = 'en' | 'es';

export interface LanguageLink {
	lang: SupportedLang;
	href: string;
	label?: string;
	available?: boolean;
	isTranslation?: boolean;
	description?: string;
}
