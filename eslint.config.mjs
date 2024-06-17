import eslintPluginYml from 'eslint-plugin-yml';

export default [
    ...eslintPluginYml.configs['flat/standard'],
    {
        files: ["docker-compose.yaml"],
        rules: {
            "yml/sort-keys": "error",
            "yml/quotes": ["error", {"prefer": "double"}],
            "yml/plain-scalar": ["error", "never"]
        }
    }
];