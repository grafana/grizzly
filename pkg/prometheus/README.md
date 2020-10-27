# Prometheus Provider for Grizzly

This provider offers a handler for Prometheus Alerts and Recording Rules, when
the Prometheus server has an API for upload, such as that provided by the
Cortex Ruler.

The default Prometheus install does not have this API, meaning Grizzly cannot
be used with it.

(Having said that, `grr export` very much could be used to generate files
that a default Prometheus install could consume - it just wouldn't be able to
be used with `grr apply`.)