schema_version = 1

project {
  license          = "MIT"
  copyright_holder = "VirtualTam"

  header_ignore = [
    # Docker Compose
    "docker-compose*.yml",

    # mdBook documentation
    "docs/**",

    # Fonts
    "**/firacode/**",
  ]
}
