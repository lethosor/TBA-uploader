module.exports = {
    "extends": [
        "eslint:recommended",
        "plugin:vue/recommended",
    ],
    "parserOptions": {
        "ecmaVersion": 2018,
        "sourceType": "module",
    },
    env: {
        "browser": true,
        "jquery": true,
    },
    "globals": {
        "FMS_CONFIG": "readonly",
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
