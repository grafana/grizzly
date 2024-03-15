const core = require('@actions/core');
const tc = require("@actions/tool-cache");
const { downloadBinary, identifyLatest } = require('./lib/utils');

const toolName = 'grr';

async function setup() {
    try {
        // Get the version to be installed
        let version = core.getInput('version');

        if (version === 'latest') {
            console.log('Identifying "latest" version...');
            version = await identifyLatest();
            console.log(`Latest version is ${version}`);
        }

        const cachedPath = tc.find(toolName, version)
        if (cachedPath) {
            // Note: this cache is only used across runs :(
            // See: https://github.com/actions/toolkit/issues/58
            console.log(`Using Grizzly ${version} from cache`);
            core.addPath(cachedPath);
            return;
        }

        const binaryDir = await downloadBinary(version);

        console.log(`Caching Grizzly ${version}`);
        const grrCacheDir = await tc.cacheDir(binaryDir, toolName, version)

        // Expose grizzly by adding it to the PATH
        console.log(`Adding Grizzly to PATH: ${grrCacheDir}`)
        core.addPath(grrCacheDir);
    } catch (e) {
        core.setFailed(e);
    }
}

module.exports = setup;

if (require.main === module) {
    setup();
}
