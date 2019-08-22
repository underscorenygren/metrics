# metrics
Experimentation with metrics related stuff in golang

### Logging

We use a global logging module approach, to make it easy
to configure in tests, without putting any burden on
people instantiating modules having to know about
the internals of logging.

This does depend on a global module, which I'm a bit iffy
about. If I see a better approach to handle this, I'd
consider using that instead.
