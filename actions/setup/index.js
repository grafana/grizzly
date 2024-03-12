const fsPromises = require('fs').promises;
const path = require('path');
const core = require('@actions/core');
const { downloadBinary, identifyLatest } = require('./lib/utils');

async function setup() {
    try {
        // Get the version to be installed
        let version = core.getInput('version');

        if (version === 'latest') {
            console.log('Identifying "latest" version...');
            version = await identifyLatest();
            console.log(`Latest version is ${version}`);
        }

        const pathToBinary = await downloadBinary(version);
        const binaryDirectory = path.dirname(pathToBinary);

        await fsPromises.rename(pathToBinary, `${binaryDirectory}/grr`);
        await fsPromises.chmod(`${binaryDirectory}/grr`, 0o755);

        // Expose grizzly by adding it to the PATH
        console.log('Adding Grizzly to PATH')
        core.addPath(binaryDirectory);
    } catch (e) {
        core.setFailed(e);
    }
}

module.exports = setup

if (require.main === module) {
    setup();
}
