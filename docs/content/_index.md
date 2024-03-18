---
date: "2021-06-28T00:00:00+00:00"
title: "Grizzly"
---
## A Command Line utility for managing Grafana Resources
Software engineers know how to version and deploy their resources. Tools like Git or CI enable
reliable workflows that track changes, with meaningful review processes giving confidence in the
expected outcomes.

Now, with **Grizzly**, you can have all this with Grafana resources, dashboards, datasources and
more.

Grizzly can be a valuable component in a number of workflows around Grafana resources.

## Edit/Publish
![Edit workflow](images/workflow-edit.png)

The first step in moving to a version controlled workflow for Grafana resources is to pull them into
files from your Grafana instance. Grizzly can help you here. You can commit these files into version
control (e.g. `git`), as you would any other file.

Then, when you make changes, the [Grizzly Server](server/) allows you to edit these files, full
wysiwyg, against a Grafana instance, without needing to publish them. Click `save` on your resource
and (only) your local file is updated, ready for your changes to be committed to git/etc.

Once correct, Grizzly can publish your resources to your Grafana instance - directly - or more likely within
a CI pipeline.

## Create/Review/Publish
![Review workflow](images/workflow-review.png)

When you create Grafana resources with code (for example with [Grafonnet](https://github.com/grafana/grafonnet)
or the [Grafana Foundation SDK](https://github.com/grafana/grafana-foundation-sdk)), confirming that your
resources are valid can be painful.

If your resources are in Jsonnet, Grizzly can process that Jsonnet for you. In any other language, just render
your resources to JSON or YAML and Grizzly will take it from there.

From here, the [Grizzly Server](server/) can be used to review and validate, and beyond that, as in the previous
workflow, Grizzly can publish those resources to your Grafana instance (likely from a CI pipeline).

## Resource Migration
![Review migrate](images/workflow-migrate.png)

Want to migrate resources between Grafana instances? Grizzly is your friend. with a combination of `grr pull`
and `grr push` you can migrate your resources e.g. from an OSS Grafana or on-prem instance to Grafana Cloud.

[Find out more](/grizzly/what-is-grizzly)
