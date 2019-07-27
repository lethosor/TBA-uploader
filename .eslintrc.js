module.exports = {
    "extends": [
        "eslint:recommended",
        "plugin:vue/recommended",
    ],
    "parserOptions": {
        "ecmaVersion": 2017,
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
    "rules": {
        "indent": [
            "error",
            4,
            {
                "MemberExpression": "off",
            },
        ],
        "no-console": "error",
        "comma-dangle": ["error", "always-multiline"],
        "semi": "error",

        "vue/html-indent": [
            "error",
            4,
        ],
        "vue/singleline-html-element-content-newline": "off",
    },
};
