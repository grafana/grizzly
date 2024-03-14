---
date: "2021-06-28T00:00:00+00:00"
title: "Workflows"
---

Grizzly supports a number of workflows.

## Static Resources
All examples so far have described Grizzly interacting with resources represented
as static resources in YAML. This is the simplest use-case for Grizzly, but there
are more powerful workflows available.

## Pull/Push
With `grr pull -d` and `grr apply -d` it is possible to migrate dashboards between
Grafana instances. To pull dashboards and folders from one instance to another
is as simple as:
```
$ grr config create-context source
$ grr config set grafana.url <...source Grafana URL...>
$ grr config set grafana.token <...source Grafana service account token...>
$ grr config set targets Dashboard,Dashboardfolder
$ grr config set output-format json
$ grr pull resources
```
This asks Grizzly to pull all resources matching the `<kind>/<UID>` pattern for
dashboards and folders into a directory called `resources`.

> **Note**: Grizzly can pull datasources, but secure passwords won't be included
> when pulled - these will need to be provided manually (either by editing into
> the downloaded YAML or pasting them in via the Grafana UI).

To push them to a new Grafana instance:
```
$ grr config create-context destination
$ grr config set grafana.url <...destination Grafana URL...>
$ grr config set grafana.token <...destination Grafana service account token...>
$ grr config set targets Dashboard,Dashboardfolder
$ grr config set output-format json
$ grr push resources
```

## Jsonnet
The most powerful workflow for Grizzly involves Jsonnet, a powerful programming
language that can be used to render JSON or YAML.

With this workflow, resources are described in Jsonnet. When Grizzly is invoked,
it evaluates the Jsonnet and renders the resources into JSON/YAML. This it can
then push to the relevant systems.

For more information see the [Jsonnet page](../jsonnet/).

# Full Command List

### grr get
Retrieves a resource from the remote system, via its UID. Its UID will be two parts separated by a dot, `<resource-type>.<resource-id>`. A dashboard might be `dashboard.mydash`:

```sh
$ grr get Dashboard.my-uid
```

### grr list
List all resources found after executing Jsonnet file.
```sh
$ grr list my-lib.libsonnet
```

### grr show
Shows the resources found after executing Jsonnet, rendered as expected for each resource type:

```sh
$ grr show my-lib.libsonnet
```

### grr diff
Compares each resource rendered by Jsonnet with the equivalent on the remote system:

```sh
$ grr diff my-lib.libsonnet
```

### grr apply
Uploads each dashboard rendered by the mixin to Grafana
```sh
$ grr apply my-lib.libsonnet
```

### grr push
"Push" is an alias for `apply`, above.

### grr watch
Watches a directory for changes. When changes are identified, the
jsonnet is executed and changes are pushed to remote systems.
The directory is watched recursively (i.e. all subdirectories are watched too),
but if new subdirectories are added, watch command needs to be re-started,
as new directories will not be picked up automatically.

This example watches the current directory for changes, then executes and applies
`my-lib.libsonnet` when changes are noticed:

```sh
$ grr watch . my-lib.libsonnet
```

### grr export
Renders Jsonnet and saves resources as files directory which is specified with
the second argument.

These resources can then be deployed This can be used with
[Grafana Provisioning](https://grafana.com/docs/grafana/latest/administration/provisioning/)
to provision dashboards that can be picked up immediately by Grafana.

```sh
$ grr export some-mixin.libsonnet my-provisioning-dir
```

## Flags

### `-t, --target strings`

It allows the targeting of resources by key, where key is in the form `<type>.<uid>`.

Targets can also be wildcards, e.g. `Dashboard.*`. If no `.` character is provided, then the target will
be matched against the resource type only (e.g. `Dashboard`). In such a case, lower case names are allowed.

Run `grr list` to get a list of resource keys in your code.

### `-J, --jpath`

It allows the targeting folder containing jsonnet library to include, should be repeated multiple times.

If not specified it include `vendor`, `lib` and local dir (`.`) folders by default.
