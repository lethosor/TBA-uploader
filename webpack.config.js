'use strict';

const path = require('path');

module.exports = {
    mode: 'development',
    entry: [
        path.join(__dirname, 'web', 'src', 'app.js'),
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
                query: {
                    presets: ['@babel/env'],
                },
            },
        ],
    },
    resolve: {
        alias: {
            src: path.join(__dirname, 'web', 'src'),
        },
    },
};
