#!/bin/bash

# Database schema verification script
DB_PATH="${DB_PATH:-.}/release_control.db"

echo "=== Database Schema Verification ==="
echo "Database: $DB_PATH"
echo ""

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
  echo "❌ Database file not found: $DB_PATH"
  exit 1
fi

echo "✓ Database file found"
echo ""

# Verify tables
TABLES=(
  "applications"
  "clusters"
  "application_cluster_configs"
  "shell_server"
  "shell_command"
  "release_record"
  "release_event"
)

echo "=== Verifying Tables ==="
for table in "${TABLES[@]}"; do
  result=$(sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='$table';")
  if [ -z "$result" ]; then
    echo "❌ Table missing: $table"
  else
    echo "✓ Table exists: $table"
    # Show column info
    sqlite3 "$DB_PATH" ".schema $table" | head -1
  fi
done

echo ""
echo "=== Application Table Schema ==="
sqlite3 "$DB_PATH" "PRAGMA table_info(applications);"

echo ""
echo "=== Cluster Table Schema ==="
sqlite3 "$DB_PATH" "PRAGMA table_info(clusters);"

echo ""
echo "=== Application-Cluster-Config Table Schema ==="
sqlite3 "$DB_PATH" "PRAGMA table_info(application_cluster_configs);"

echo ""
echo "=== Shell Command Table Schema ==="
sqlite3 "$DB_PATH" "PRAGMA table_info(shell_command);"

echo ""
echo "=== Record Counts ==="
sqlite3 "$DB_PATH" "SELECT 'applications: ' || COUNT(*) FROM applications;"
sqlite3 "$DB_PATH" "SELECT 'clusters: ' || COUNT(*) FROM clusters;"
sqlite3 "$DB_PATH" "SELECT 'application_cluster_configs: ' || COUNT(*) FROM application_cluster_configs;"
sqlite3 "$DB_PATH" "SELECT 'shell_command: ' || COUNT(*) FROM shell_command;"
sqlite3 "$DB_PATH" "SELECT 'release_record: ' || COUNT(*) FROM release_record;"

echo ""
echo "=== Verification Complete ==="
