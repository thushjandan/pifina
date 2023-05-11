import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		proxy: {
			'/api': {
				target: 'https://localhost:8655',
				secure: false
			},
		},
		fs: {
			allow: [
				'..'
			]
		}
	}
});
