import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightVersions from 'starlight-versions';

// https://astro.build/config
export default defineConfig({
	site: 'https://mistermoe.github.io',
	base: import.meta.env.DEV ? '/' : '/httpr',
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
				},
				{
					label: 'Query Params',
					link: '/query-params'
				},
				{
					label: 'Inspect Request',
					link: '/inspect'
				}
			],
			plugins: [
				starlightVersions({
          versions: [{ slug: '1.0' }],
        }),
			]
		}),
	],
});
