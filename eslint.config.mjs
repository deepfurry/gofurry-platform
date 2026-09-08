import js from '@eslint/js';
import ts from 'typescript-eslint';
import astro from 'eslint-plugin-astro';
import hooks from 'eslint-plugin-react-hooks';

export default ts.config(
  { ignores: ['**/node_modules/**', '**/dist/**', '**/.astro/**', '**/generated/**', 'server/**', '.local/**', '.cache/**'] },
  js.configs.recommended,
  ...ts.configs.recommended,
  ...astro.configs['flat/recommended'],
  { files: ['**/*.{js,mjs,ts,tsx}'], languageOptions: { globals: {
    process: 'readonly', console: 'readonly', URL: 'readonly', fetch: 'readonly',
    document: 'readonly', window: 'readonly', Request: 'readonly', Response: 'readonly',
    setTimeout: 'readonly', clearTimeout: 'readonly', AbortController: 'readonly',
  } } },
  { files: ['**/*.tsx'], plugins: { 'react-hooks': hooks }, rules: {
    'react-hooks/rules-of-hooks': 'error', 'react-hooks/exhaustive-deps': 'error',
  } },
);
