interface OpenGraphOptions {
	title?: string;
	description?: string;
	image?: string;
	url?: string;
	type?: 'website' | 'article' | 'profile';
	siteName?: string;
}

export const defaultOpenGraph: Required<OpenGraphOptions> = {
	title: 'Mount Moreland Night Owls',
	description: 'Community watch scheduling and incident reporting for Mount Moreland',
	image: '/logo.png',
	url: '',
	type: 'website',
	siteName: 'Mount Moreland Night Owls'
};

/**
 * Generate OpenGraph meta tag strings for use in svelte:head
 * @param options Partial options to override defaults
 * @returns Object with formatted meta tag strings
 */
export function generateOpenGraphTags(options: OpenGraphOptions = {}) {
	const og = { ...defaultOpenGraph, ...options };

	return {
		// Essential OpenGraph tags
		ogType: `<meta property="og:type" content="${og.type}" />`,
		ogTitle: `<meta property="og:title" content="${og.title}" />`,
		ogDescription: `<meta property="og:description" content="${og.description}" />`,
		ogImage: `<meta property="og:image" content="${og.image}" />`,
		ogImageAlt: `<meta property="og:image:alt" content="${og.title} Logo" />`,
		ogUrl: og.url ? `<meta property="og:url" content="${og.url}" />` : '',
		ogSiteName: `<meta property="og:site_name" content="${og.siteName}" />`,

		// Twitter Card tags
		twitterCard: `<meta name="twitter:card" content="summary_large_image" />`,
		twitterTitle: `<meta name="twitter:title" content="${og.title}" />`,
		twitterDescription: `<meta name="twitter:description" content="${og.description}" />`,
		twitterImage: `<meta name="twitter:image" content="${og.image}" />`,
		twitterImageAlt: `<meta name="twitter:image:alt" content="${og.title} Logo" />`,

		// Basic SEO
		title: og.title,
		description: `<meta name="description" content="${og.description}" />`
	};
}

/**
 * Common OpenGraph configurations for different page types
 */
export const openGraphPresets = {
	home: {
		title: 'Mount Moreland Night Owls - Community Watch',
		description:
			'Join our community watch program. Schedule patrol shifts, report incidents, and help keep Mount Moreland safe.',
		type: 'website' as const
	},

	admin: {
		title: 'Night Owls Admin - Community Management',
		description:
			'Administrative dashboard for managing community watch operations, schedules, and reports.',
		type: 'website' as const
	},

	shifts: {
		title: 'Patrol Shifts - Mount Moreland Night Owls',
		description: 'View and book patrol shifts to help keep our community safe.',
		type: 'website' as const
	},

	reports: {
		title: 'Incident Reports - Mount Moreland Night Owls',
		description: 'Community incident reports and safety information.',
		type: 'website' as const
	},

	login: {
		title: 'Sign In - Mount Moreland Night Owls',
		description: 'Sign in to access your Night Owls account and community watch features.',
		type: 'website' as const
	},

	register: {
		title: 'Join Us - Mount Moreland Night Owls',
		description:
			'Join the Mount Moreland Night Owls community watch program and help keep our neighborhood safe.',
		type: 'website' as const
	}
};

/**
 * Helper to get OpenGraph tags for a specific page type
 */
export function getPageOpenGraph(
	pageType: keyof typeof openGraphPresets,
	customOptions: OpenGraphOptions = {}
) {
	const preset = openGraphPresets[pageType];
	return generateOpenGraphTags({ ...preset, ...customOptions });
}
