import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	integrations: [
		starlight({
			title: 'httpr',
			social: {
				github: 'https://github.com/mistermoe/httpr',
			},
			sidebar: [],
		}),
	],
});
