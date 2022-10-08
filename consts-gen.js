const fs = require('fs');

const commandLineArgs = require('command-line-args');
const JSON5 = require('json5');

const GENERATED_WARNING = '//\n// This file was generated. DO NOT EDIT!\n//\n\n';

function generateGo(consts, { package = 'main' }) {
    return GENERATED_WARNING + 'package ' + package + '\n\n' + Object.entries(consts).map(([type, values]) =>
        'const (\n' + Object.entries(values).map(([key, value]) =>
            `\t${type}_${key} = ${JSON.stringify(value)}\n`
        ).join('') + ')\n\n'
    ).join('');
}

function generateJs(consts) {
    return GENERATED_WARNING + Object.entries(consts).map(([type, values]) =>
        `export const ${type} = Object.freeze(${JSON.stringify(values)});\n`
    ).join('');
}

function main() {
    const args = commandLineArgs([
        {name: 'input-file', type: String, defaultOption: true},
        {name: 'output-go', type: String},
        {name: 'go-package', type: String},
        {name: 'output-js', type: String},
    ]);

    if (!args['input-file']) {
        throw new Error('Missing input file');
    }

    const inJson = JSON5.parse(fs.readFileSync(args['input-file']));
    if (args['output-go']) {
        fs.writeFileSync(args['output-go'], generateGo(inJson, {
            package: args['go-package'],
        }));
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
