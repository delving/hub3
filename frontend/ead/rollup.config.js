import preprocess from 'svelte-preprocess';
import svelte from 'rollup-plugin-svelte';
import commonjs from '@rollup/plugin-commonjs';
import resolve from '@rollup/plugin-node-resolve';
import json from '@rollup/plugin-json';
import livereload from 'rollup-plugin-livereload';
import {terser} from 'rollup-plugin-terser';
import scss from 'rollup-plugin-scss';
import {string} from 'rollup-plugin-string'
import serve from 'rollup-plugin-serve'

const customerId = 'na';
const production = !process.env.ROLLUP_WATCH;

export default {
	input: 'src/main.js',
	output: {
		sourcemap: true,
		format: 'esm',
		name: 'app',
		dir: 'public/build/'
	},
	plugins: [
    serve({
      contentBase: ['public'],
      port: 5000,
      historyApiFallback: `/${customerId}/index.html`
    }),
    json(),

		svelte({
			compilerOptions: {
				// enable run-time checks when not in production
				dev: !production,
			},
      preprocess: preprocess()
		}),
    string({
      include: '**/*.xml'
    }),
		resolve({
			browser: true,
			dedupe: ['svelte']
		}),
    // we'll extract any component CSS out into
		// a separate file - better for performance

		// If you have external dependencies installed from
		// npm, you'll most likely need these plugins. In
		// some cases you'll need additional configuration -
		// consult the documentation for details:
		// https://github.com/rollup/plugins/tree/master/packages/commonjs
    commonjs(),
    scss({ output: 'public/build/bundle.css' }),

		// Watch the `public` directory and refresh the
		// browser on changes when not in production
		!production && livereload('public'),

		// If we're building for production (npm run build
		// instead of npm run dev), minify
		production && terser()
	],
	watch: {
		clearScreen: false
	}
};
