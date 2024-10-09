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
				'@fontsource/lato/index.css',
				'@fontsource/lato/100.css',
				'@fontsource/lato/100-italic.css',
				'@fontsource/lato/300.css',
				'@fontsource/lato/300-italic.css',
				'@fontsource/lato/700.css',
				'@fontsource/lato/700-italic.css',
				'@fontsource/lato/900.css',
				'@fontsource/lato/900-italic.css',
				'@fontsource/lato/latin.css',
				'@fontsource/lato/latin-ext.css',
				'@fontsource/lato/latin-italic.css',
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
					label: 'Request Body',
					link: '/request-body'
				},
				{
					label: 'Response Body',
					link: '/response-body'
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
