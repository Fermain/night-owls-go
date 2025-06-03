export default {
	'*.{js,ts,svelte}': ['eslint --fix', 'prettier --write'],
	'*.{json,md,yaml,yml}': ['prettier --write'],
	'*.svelte': ['svelte-check --tsconfig ./tsconfig.json']
};
