const path = require('path');
const fsPromises = require('fs').promises;
const os = require('os');
const tc = require("@actions/tool-cache");

// arch in [arm, x32, x64...] (https://nodejs.org/api/os.html#os_os_arch)
// return value in [amd64, arm, arm64, ...]
function mapArch(arch) {
    const mappings = {
        x64: 'amd64'
    };
    return mappings[arch] || arch;
}

// os in [darwin, linux, win32...] (https://nodejs.org/api/os.html#os_os_platform)
// return value in [darwin, linux, windows]
function mapOS(os) {
    const mappings = {
        win32: 'windows'
    };
    return mappings[os] || os;
}

function getDownloadURL(version) {
    const filename = `grr-${mapOS(os.platform())}-${mapArch(os.arch())}`;

    return `https://github.com/grafana/grizzly/releases/download/${version}/${filename}`;
}

async function downloadBinary(version) {
    // Download the specific version of grizzly as a binary
    const binaryURL = getDownloadURL(version);

    console.log(`Downloading Grizzly ${version} from ${binaryURL}`);

    const pathToBinary = await tc.downloadTool(binaryURL);
    const binaryDir = path.dirname(pathToBinary);

    await fsPromises.chmod(pathToBinary, 0o755);
    await fsPromises.rename(pathToBinary, `${binaryDir}/grr`);

    return binaryDir;
}

async function identifyLatest() {
    const response = await fetch('https://api.github.com/repos/grafana/grizzly/releases/latest');
    const release = await response.json()

    return release.tag_name;
}

module.exports = {
    downloadBinary,
    identifyLatest,
}
