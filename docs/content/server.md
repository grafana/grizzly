---
date: 2024-03-14
title: Grizzly Server
---
## An HTTP Server for editing and reviewing
When we manage Grafana dashboards on disk, the "Grizzly Server" makes easy to
edit and review these resources within an actual Grafana instance, but without
needing to publish the dashboard to Grafana. Grafana provides the UI and datasources
to make dashboards look right, Grizzly provides the dashboard, directly from
local disk.

The Grizzly server is configured the same way as other Grizzly commands. See
the [configuration section](../configuration)) for more details.

### Starting the Grizzly Server
The Grizzly server can be started with this command:

```
grr serve
```

This will start a server, by default listening on port 8080, but it will not have any resources available
for you. It can be opened at http://localhost:8080. With the `-b` argument, Grizzly will open a browser
window. With `-p`, you can change the port that the Grizzly server listens on.

The next sections will explain the four main scenarios for which the Grizzly server is useful.

### Editing JSON or YAML files in Grafana
You can run Grizzly against one or more local files and it will start up an
HTTP server:

```
grr serve <mydir>
```

By default, visit http://localhost:8080 to view the Grizzly server.

### Reviewing changes to JSON or YAML files in Grafana
If you are editing the resources on disk, and just want to use Grafana for review, then use the inbuilt
"watch" functionality. With the below, if any files are changed on disk within the directory identified,
(here, `mydir`), the dashboard will be reloaded within Grafana:

```
grr serve -w <mydir>
```

This could be useful if, for example, you use another language (other than jsonnet) to render your
JSON/YAML and want to see the outcomes in Grafana.

### Reviewing changes to your Jsonnet scripts in Grafana
If you are working with Jsonnet, and your jsonnet codebase covers more than one file, you can specify
the entrypoint for your Jsonnet and the directory to watch independently:

```
grr serve -w <my-jsonnet-script> <watch-dir>
```

### Reviewing changes to code in other languages in Grafana
The [Grafana Foundation SDK](https://github.com/grafana/grafana-foundation-sdk) provides libraries in a
range of languages that can be used to render Grafana dashboards. Watching changes to these with Grizzly
is a two stage process, currently requiring an additional tool to watch for changes to source code and
render your dashboard(s) to files. One such tool is [entr](https://github.com/eradman/entr), which can be
used like so (with the Foundation SDK's TypeScript support):

```
cd grafana-foundation-sdk/examples/typescript/red-method
npm install
find . | entr -s 'npm run -s dev > ts.json'
```
Then, in another window:
```
grr serve -w ts.json
```
