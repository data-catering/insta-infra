import eslintPluginYml from 'eslint-plugin-yml';

export default [
    ...eslintPluginYml.configs['flat/standard'],
    {
        files: ["docker-compose.yaml"],
        rules: {
            "yml/sort-keys": "error"
        }
    }
];