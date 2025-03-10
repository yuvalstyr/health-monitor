env "local" {
  url = "sqlite://health-monitor.db"
  dev = "sqlite://health-monitor.db"
  migration {
    dir = "file://migrations"
  }
}

lint {
  latest = "all"
}
