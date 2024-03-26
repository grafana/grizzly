const spawn = require('child_process').spawn;

function setConfig(key, value, io = {}) {
    return new Promise(function (resolve, reject) {
        const process = spawn('grr', ['config', 'set', key, value]);
        process.on('close', function (code) {
            resolve(code);
        });
        process.on('error', function (err) {
            throw err;
        });

        if (io.stdout) {
            process.stdout.on('data', (data) => {
                stdout(`${data}`);
            });
        }
        if (io.stderr) {
            process.stderr.on('data', (data) => {
                stderr(`${data}`);
            });
        }
    });
}

async function writeConfig(config, io = {}) {
    for (const [key, value] of Object.entries(config)) {
        const exitCode = await setConfig(key, value, io);
        if (exitCode !== 0) {
            throw new Error(`could not set config key ${key}`);
        }
    }
}

module.exports = {
    writeConfig,
}
