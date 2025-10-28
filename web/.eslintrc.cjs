module.exports = {
    root: true,
    env: { browser: true, es2020: true },
    extends: [
      'eslint:recommended',
      'plugin:@typescript-eslint/recommended',
      'plugin:react-hooks/recommended',
      'airbnb', // Extends Airbnb's base rules
      'airbnb/hooks', // Extends Airbnb's React Hooks rules
      'airbnb-typescript', // Extends Airbnb's rules for TypeScript
    ],
    ignorePatterns: ['dist', '.eslintrc.cjs', 'vite.config.js'],
    parser: '@typescript-eslint/parser',
    parserOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      // Point ESLint to your tsconfig.json files for TypeScript-aware linting
      project: ['./tsconfig.json', './tsconfig.node.json'],
      tsconfigRootDir: __dirname, // Specify the root directory for tsconfig.json
    },
    plugins: ['react-refresh'],
    rules: {
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
      // Disable prop-types as we're using TypeScript
      'react/prop-types': 'off',
      // Allow .tsx and .jsx extensions
      'react/jsx-filename-extension': [1, { extensions: ['.js', '.jsx', '.ts', '.tsx'] }],
      // Relax rules for default exports (common in React components)
      'import/prefer-default-export': 'off',
      // No-shadow can conflict with TypeScript enums and types, use TS version
      'no-shadow': 'off',
      '@typescript-eslint/no-shadow': ['error'],
      // Disable 'no-use-before-define' for functions and types for better organization
      'no-use-before-define': 'off',
      '@typescript-eslint/no-use-before-define': ['error', { functions: false, classes: true, variables: true, typedefs: true }],
      // Allow explicit any if necessary (can be adjusted to 'error' later)
      '@typescript-eslint/no-explicit-any': 'warn',
      // Allow unused vars for arguments that are prefixed with an underscore
      '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
      // Enforce consistent use of type imports (e.g., `import type { Foo } from 'bar'`)
      "@typescript-eslint/consistent-type-imports": "error",
    },
  };
  