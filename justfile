set unstable
set dotenv-load

# List all available recipes
default:
    just -l -u

# Run the app
run:
    go run cmd/attendance-tracker/main.go

# Build the app
build:
    go build -o attendance-tracker cmd/attendance-tracker/main.go

# Run the app after generating templates and css styles
run_gen: gogen tailwindcss
    go run cmd/attendance-tracker/main.go

# Build the app after generating templates and css styles
build_gen: gogen tailwindcss
    go build -o attendance-tracker cmd/attendance-tracker/main.go

# Run all migrations
migrate_up_all:
    goose up

# Run the next migration
migrate_up_one:
    goose up-by-one

# Run all migrations up to the specified version
migrate_up_to version:
    goose up-to {{version}}

# Rollback a migration
migrate_down_one:
    goose down

# Rollback migrations to specified version
migrate_down_to version:
    goose down-to {{version}}

# Create a new migration with the specified name
migrate_create name:
    goose -s create {{name}} sql

# Generate database-facing code using sqlc
gensql:
    sqlc generate

# Call go generate to automate parts of the build process
gogen:
    go generate ./...

# Call tailwind CLI to generate styles
[working-directory: 'static']
tailwindcss:
    npx tailwindcss -i static/style.css -o static/css/style.css -m --cwd ..
