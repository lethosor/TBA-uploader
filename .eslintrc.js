module.exports = {
    "extends": "eslint:recommended",
    "parserOptions": {
        "ecmaVersion": 6,
        "sourceType": "module",
    },
    env: {
        "browser": true,
        "jquery": true,
    },
    "globals": {
        "Vue": "readonly",
        "showdown": "readonly",
        "MATCH_LEVEL_QUAL": "readonly",
        "MATCH_LEVEL_PLAYOFF": "readonly",
    },
};
