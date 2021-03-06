const fs = require('fs');

const commandLineArgs = require('command-line-args');
const JSON5 = require('json5');

function generateGo(consts) {
    return 'package main\n\n' + Object.entries(consts).map(([type, values]) =>
        'const (\n' + Object.entries(values).map(([key, value]) =>
            `\t${type}_${key} = ${JSON.stringify(value)}\n`
        ).join('') + ')\n\n'
    ).join('');
}

function generateJs(consts) {
    return Object.entries(consts).map(([type, values]) =>
        `export const ${type} = Object.freeze(${JSON.stringify(values)});\n`
    ).join('');
}

function main() {
    const args = commandLineArgs([
        {name: 'input-file', type: String, defaultOption: true},
        {name: 'output-go', type: String},
        {name: 'output-js', type: String},
    ]);

    if (!args['input-file']) {
        throw new Error('Missing input file');
    }

    const inJson = JSON5.parse(fs.readFileSync(args['input-file']));
    if (args['output-go']) {
        fs.writeFileSync(args['output-go'], generateGo(inJson));
    }
    if (args['output-js']) {
        fs.writeFileSync(args['output-js'], generateJs(inJson));
    }
}

try {
    main();
}
catch (e) {
    console.error(e.message || e);
    process.exit(1);
}
