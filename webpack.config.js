'use strict';

const path = require('path');

module.exports = {
    mode: 'development',
    entry: [
        path.join(__dirname, 'web', 'app.js'),
    ],
    output: {
        filename: 'bundle.js',
        path: path.join(__dirname, 'web'),
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
};
