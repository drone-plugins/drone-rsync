local windows_pipe = '\\\\\\\\.\\\\pipe\\\\docker_engine';
local windows_pipe_volume = 'docker_pipe';
local test_pipeline_name = 'testing';

local windows(os) = os == 'windows';

local golang_image(os, version) =
  if windows(os) then 'golang:1.11-windowsservercore-' + version
  else 'golang:1.26.10-alpine@sha256:27f2d3b83cfc5803c8078f8e730af0f7c431bf067afaec8f60a36fb763187cba';

local setup_go(os) = if windows(os) then ['. ./scripts/setup-go-windows.ps1'] else [];

{
  test(os='linux', arch='amd64', version='')::
    local is_windows = windows(os);
    local golang = golang_image(os, version);
    local volumes = if is_windows then [{name: 'gopath', path: 'C:\\\\gopath'}] else [{name: 'gopath', path: '/go',}];
    {
      kind: 'pipeline',
      name: test_pipeline_name,
      platform: {
        os: os,
        arch: arch,
        version: if std.length(version) > 0 then version,
      },
      steps: [
        {
          name: 'vet',
          image: golang,
          pull: 'always',
          environment: {
            GO111MODULE: 'on',
          },
          commands: setup_go(os) + [
            'go vet ./...',
          ],
          volumes: volumes,
        },
        {
          name: 'test',
          image: golang,
          pull: 'always',
          environment: {
            GO111MODULE: 'on',
          },
          commands: setup_go(os) + [
            'go test -cover ./...',
          ],
          volumes: volumes,
        },
      ],
      trigger: {
        ref: [
          'refs/heads/master',
          'refs/tags/**',
          'refs/pull/**',
        ],
      },
      volumes: [{name: 'gopath', temp: {}}]
    },

  build(name, os='linux', arch='amd64', version='')::
    local is_windows = windows(os);
    local tag = if is_windows then os + '-' + version else os + '-' + arch;
    local file_suffix = std.strReplace(tag, '-', '.');
    local volumes = if is_windows then [{ name: windows_pipe_volume, path: windows_pipe }] else [];
    local golang = golang_image(os, version);
    local plugin_repo = 'plugins/' + std.splitLimit(name, '-', 1)[1];
    local extension = if is_windows then '.exe' else '';
    {
      kind: 'pipeline',
      name: tag,
      platform: {
        os: os,
        arch: arch,
        version: if std.length(version) > 0 then version,
      },
      steps: [
        {
          name: 'build-push',
          image: golang,
          pull: 'always',
          environment: {
            CGO_ENABLED: '0',
            GO111MODULE: 'on',
          },
          commands: setup_go(os) + [
            'go build -v -ldflags "-X main.version=${DRONE_COMMIT_SHA:0:8}" -a -tags netgo -o release/' + os + '/' + arch + '/' + name + extension,
          ],
          when: {
            event: {
              exclude: ['tag'],
            },
          },
        },
        {
          name: 'build-tag',
          image: golang,
          pull: 'always',
          environment: {
            CGO_ENABLED: '0',
            GO111MODULE: 'on',
          },
          commands: setup_go(os) + [
            'go build -v -ldflags "-X main.version=${DRONE_TAG##v}" -a -tags netgo -o release/' + os + '/' + arch + '/' + name + extension,
          ],
          when: {
            event: ['tag'],
          },
        },
        {
          name: 'executable',
          image: golang,
          pull: 'always',
          commands: [
            './release/' + os + '/' + arch + '/' + name + extension + ' --help',
          ],
        },
        {
          name: 'dryrun',
          image: 'plugins/docker:' + tag,
          pull: 'always',
          settings: {
            dry_run: true,
            tags: tag,
            dockerfile: 'docker/Dockerfile.' + file_suffix,
            daemon_off: if is_windows then 'true' else 'false',
            repo: plugin_repo,
            username: { from_secret: 'docker_username' },
            password: { from_secret: 'docker_password' },
          },
          volumes: if std.length(volumes) > 0 then volumes,
          when: {
            event: ['pull_request'],
          },
        },
        {
          name: 'publish',
          image: 'plugins/docker:' + tag,
          pull: 'always',
          settings: {
            auto_tag: true,
            auto_tag_suffix: tag,
            daemon_off: if is_windows then 'true' else 'false',
            dockerfile: 'docker/Dockerfile.' + file_suffix,
            repo: plugin_repo,
            username: { from_secret: 'docker_username' },
            password: { from_secret: 'docker_password' },
          },
          volumes: if std.length(volumes) > 0 then volumes,
          when: {
            event: {
              exclude: ['pull_request'],
            },
          },
        },
      ],
      trigger: {
        ref: [
          'refs/heads/master',
          'refs/tags/**',
          'refs/pull/**',
        ],
      },
      depends_on: [test_pipeline_name],
      volumes: if is_windows then [{ name: windows_pipe_volume, host: { path: windows_pipe } }],
    },

  notifications(os='linux', arch='amd64', version='', depends_on=[])::
    {
      kind: 'pipeline',
      name: 'notifications',
      platform: {
        os: os,
        arch: arch,
        version: if std.length(version) > 0 then version,
      },
      steps: [
        {
          name: 'manifest',
          image: 'plugins/manifest',
          pull: 'always',
          settings: {
            username: { from_secret: 'docker_username' },
            password: { from_secret: 'docker_password' },
            spec: 'docker/manifest.tmpl',
            ignore_missing: true,
            auto_tag: true,
          },
        },
        {
          name: 'microbadger',
          image: 'plugins/webhook',
          pull: 'always',
          settings: {
            urls: { from_secret: 'microbadger_url' },
          },
        },
      ],
      trigger: {
        ref: [
          'refs/heads/master',
          'refs/tags/**',
        ],
      },
      depends_on: depends_on,
    },
}
