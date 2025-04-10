import eslintPluginYml from 'eslint-plugin-yml';

export default [
    ...eslintPluginYml.configs['flat/standard'],
    {
        files: ["cmd/insta/resources/docker-compose.yaml", "cmd/insta/resources/docker-compose-persist.yaml"],
        rules: {
            "yml/sort-keys": "error"
        }
    }
];