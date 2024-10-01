import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://mistermoe.github.io',
	base: '/httpr',
	integrations: [
		starlight({
			title: 'httpr',
			customCss: [
				'@fontsource-variable/roboto-mono/index.css',
				'./src/styles/custom.css'
			],
			social: {
				github: 'https://github.com/mistermoe/httpr',
			},
			sidebar: [
				{ 
					label: 'Getting Started',
					link: '/'
				},
				{
					label: 'Headers',
					link: '/headers'
				}
			],
		}),
	],
});
