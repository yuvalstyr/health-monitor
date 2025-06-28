env "local" {
  url = "sqlite://health.db"
  dev = "sqlite://health.db"
  migration {
    dir = "file://migrations"
  }
}

lint {
  latest = "all"
}
