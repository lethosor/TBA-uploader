'use strict';

const path = require('path');

const VueLoaderPlugin = require('vue-loader/lib/plugin');

module.exports = {
    mode: 'development',
    plugins: [
        new VueLoaderPlugin(),
    ],
    entry: [
        path.join(__dirname, 'web', 'src', 'main.js'),
    ],
    output: {
        filename: 'bundle.js',
        path: path.join(__dirname, 'web', 'dist'),
    },
    module: {
        rules: [
            {
                test: /\.js$/,
                loader: 'babel-loader',
                options: {
                    presets: ['@babel/env'],
                    overrides: [{
                        test: './node_modules/bootstrap-vue/esm/icons/icons.js',
                        compact: true,
                    }],
                },
            },
            {
                test: /\.vue$/,
                loader: 'vue-loader',
            },
            {
                test: /\.css$/,
                use: [
                    'vue-style-loader',
                    'css-loader',
                ],
            },
        ],
    },
    resolve: {
        alias: {
            src: path.join(__dirname, 'web', 'src'),
            components: path.join(__dirname, 'web', 'src', 'components'),
        },
    },
};
