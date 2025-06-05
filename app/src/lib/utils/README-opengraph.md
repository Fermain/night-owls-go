# OpenGraph Implementation

This directory contains utilities for managing OpenGraph meta tags across the Night Owls app. OpenGraph tags control how pages appear when shared on social media platforms like Facebook, Twitter, LinkedIn, etc.

## Base Configuration

The base OpenGraph configuration is set in `app/src/app.html` and includes:

- **Image**: `/logo.png` (the Night Owls logo)
- **Title**: "Mount Moreland Night Owls"
- **Description**: "Community watch scheduling and incident reporting for Mount Moreland"
- **Type**: website
- **URL**: Dynamic based on current page

## Page-Specific OpenGraph

For pages that need custom OpenGraph tags, use the utility in `opengraph.ts`:

### Quick Usage

```svelte
<script lang="ts">
	import { getPageOpenGraph } from '$lib/utils/opengraph';
	// ... other imports
</script>

{@const ogTags = getPageOpenGraph('home')}

<svelte:head>
	<title>{ogTags.title}</title>
	{@html ogTags.description}
	{@html ogTags.ogTitle}
	{@html ogTags.ogDescription}
	{@html ogTags.ogImage}
	{@html ogTags.ogImageAlt}
	{@html ogTags.ogType}
	{@html ogTags.ogSiteName}
	{@html ogTags.twitterCard}
	{@html ogTags.twitterTitle}
	{@html ogTags.twitterDescription}
	{@html ogTags.twitterImage}
	{@html ogTags.twitterImageAlt}
</svelte:head>
```

### Available Presets

The following page types have predefined configurations:

- `home` - Main dashboard/landing page
- `admin` - Admin area pages
- `shifts` - Shift scheduling pages
- `reports` - Incident reports pages
- `login` - Sign in page
- `register` - Account creation page

### Custom Configuration

For pages needing unique OpenGraph data:

```svelte
<script lang="ts">
	import { generateOpenGraphTags } from '$lib/utils/opengraph';

	const ogTags = generateOpenGraphTags({
		title: 'Custom Page Title',
		description: 'Custom description for this specific page',
		type: 'article' // or 'website', 'profile'
	});
</script>

<svelte:head>
	<title>{ogTags.title}</title>
	<!-- ... include other meta tags as needed -->
</svelte:head>
```

## Implementation Notes

1. **Logo Image**: All pages use the logo at `/logo.png` by default
2. **Fallback**: Pages without specific OpenGraph tags will use the base configuration from `app.html`
3. **Twitter Cards**: Automatically generated to match OpenGraph settings
4. **Image Dimensions**: Logo is set with recommended dimensions (1200x630) for optimal display

## Testing

To test OpenGraph implementation:

1. **Facebook Debugger**: https://developers.facebook.com/tools/debug/
2. **Twitter Card Validator**: https://cards-dev.twitter.com/validator
3. **LinkedIn Post Inspector**: https://www.linkedin.com/post-inspector/

Simply paste your page URL to see how it will appear when shared.

## Examples

Current implementations:

- **Home Page** (`/`): Uses `home` preset
- **Login Page** (`/login`): Uses `login` preset

To add OpenGraph to new pages, follow the pattern shown in these examples.
