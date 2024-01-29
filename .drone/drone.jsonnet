local golang = 'golang:1.20.6';

local volumes = [{ name: 'docker', host: { path: '/var/run/docker.sock' } }, { name: 'gopath', temp: {} }];
local mounts = [{ name: 'gopath', path: '/go' }, { name: 'docker', path: '/var/run/docker.sock' }];

local constraints = {
  onlyTagOrMain: { trigger: {
    ref: [
      'refs/heads/main',
      'refs/heads/docker',
      'refs/tags/v*',
    ],
  } },
  onlyTags: { trigger: {
    event: ['tag'],
  } },
  always: {},
};

local go(name, commands) = {
  name: name,
  image: golang,
  volumes: mounts,
  commands: commands,
};

local make(target) = go(target, ['make ' + target]);

local pipeline(name) = {
  kind: 'pipeline',
  name: name,
  type: 'docker',
  volumes: volumes,
  steps: [],
};

local docker(arch) = pipeline('docker-' + arch) {
  platform: {
    os: 'linux',
    arch: arch,
  },
  steps: [
    make('static'),
    {
      name: 'container',
      image: 'plugins/docker',
      settings: {
        repo: 'grafana/grizzly',
        auto_tag: true,
        auto_tag_suffix: arch,
        username: { from_secret: 'docker_username' },
        password: { from_secret: 'docker_password' },
      },
    },
  ],
};

local vault_secret(name, vault_path, key) = {
  kind: 'secret',
  name: name,
  get: {
    path: vault_path,
    name: key,
  },
};

[
  pipeline('release') {
    steps: [
      go('fetch-tags', ['git fetch origin --tags']),
      make('cross'),
      {
        name: 'publish',
        image: 'plugins/github-release',
        settings: {
          title: '${DRONE_TAG}',
          note: importstr 'release-note.md',
          api_key: { from_secret: 'github_token' },
          files: 'dist/*',
          draft: true,
        },
      },
    ],
  } + constraints.onlyTags,

  docker('amd64') + constraints.onlyTagOrMain,
  docker('arm64') + constraints.onlyTagOrMain,

  pipeline('manifest') {
    steps: [{
      name: 'manifest',
      image: 'plugins/manifest',
      settings: {
        auto_tag: true,
        ignore_missing: true,
        spec: '.drone/docker-manifest.tmpl',
        username: { from_secret: 'docker_username' },
        password: { from_secret: 'docker_password' },
      },
    }],
  } + {
    depends_on: [
      'docker-amd64',
      'docker-arm64',
    ],
  } + constraints.onlyTagOrMain,
]
+ [
  vault_secret('github_token', 'infra/data/ci/github/grafanabot', 'pat'),
  vault_secret('docker_username', 'infra/data/ci/docker_hub', 'username'),
  vault_secret('docker_password', 'infra/data/ci/docker_hub', 'password'),
]
