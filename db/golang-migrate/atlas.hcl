data "external_schema" "gorm" {
  program = ["./atlas-loader"]
}
env "gorm" {
  src = data.external_schema.gorm.url
  migration {
    dir = "file://migrations"
    format = golang-migrate
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}