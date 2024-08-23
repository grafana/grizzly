local dashboard(uid, title) = {
  uid: uid, 
  title: title,
  tags: ['templated'],
  timezone: 'browser',
  schemaVersion: 17,
  panels: [],
};

[
  dashboard("dashboard-%d" % num, "Dashboard %d" % num)
  for num in std.range(1,50)
]
