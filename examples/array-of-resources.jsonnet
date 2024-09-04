local dashboard(uid, title) = {
  uid: uid,
  title: title,
  tags: ['templated'],
  timezone: 'browser',
  schemaVersion: 17,
  panels: [],
};

[
  dashboard("dashboard-1", "Dashboard 1"),
  dashboard("dashboard-2", "Dashboard 2"),
  dashboard("dashboard-3", "Dashboard 3"),
]
