schema_version = 1

project {
  license          = "MIT"
  copyright_holder = "VirtualTam"

  header_ignore = [
    # IDEs
    ".idea/**",

    # Docker Compose
    "docker-compose*.yml",

    # mdBook documentation
    "docs/**",

    # Fonts
    "**/firacode/**",

    # Generated assets
    "internal/http/www/assets/css/chroma.css",
    "internal/http/www/static/*.min.css",
    "internal/http/www/static/*.min.js",
  ]
}
