# 🚀 go-tecton - Next-Generation Database Migration & Seeding CLI

[![Go Version](https://img.shields.io/github/go-mod/go-version/fresmaa/go-tecton)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/fresmaa/go-tecton)](https://goreportcard.com/report/github.com/fresmaa/go-tecton)
[![Release](https://img.shields.io/github/v/release/fresmaa/go-tecton)](https://github.com/fresmaa/go-tecton/releases)

`go-tecton` is an enterprise-grade database migration tool built for modern engineering teams. Born out of the frustration with cryptic error messages and "dirty database states", `go-tecton` brings safety, static analysis, and world-class Developer Experience (DX) to your schema management right inside your terminal.

---

## 🆚 Why go-tecton? (Head-to-Head Comparison)

While tools like `golang-migrate` and `goose` are great at executing SQL, they lack modern Quality Assurance (QA) and Developer Experience (DX) features. `go-tecton` isn't just a runner; it's a safety net for your database.

| Feature                                 | `go-tecton` |     `golang-migrate`      |        `goose`        |
| :-------------------------------------- | :---------: | :-----------------------: | :-------------------: |
| **Anti-Dirty State (Transactional)**    |   🟢 Yes    | 🔴 No (Manual fix needed) |        🟢 Yes         |
| **Visual Error DX (Line Highlighting)** |   🟢 Yes    |   🔴 No (Raw DB errors)   | 🔴 No (Raw DB errors) |
| **Static SQL Linter (Pre-flight QA)**   |   🟢 Yes    |           🔴 No           |         🔴 No         |
| **Dry-Run Engine (Simulate first)**     |   🟢 Yes    |           🔴 No           |      🟡 Partial       |
| **Dedicated Stateless Seeder**          |   🟢 Yes    |           🔴 No           |         🔴 No         |

---

## ✨ Key Innovations

- 🛡️ **Anti-Dirty State Engine:** Migrations are wrapped in strict transactions. If a query fails halfway, the engine automatically rolls back. Your database will never be left in a broken, unrecoverable state again.
- 🎨 **Beautiful Visual Error DX:** Say goodbye to generic SQL syntax errors. `go-tecton` pinpoints the exact line of your broken SQL file, rendering a beautiful error box in your terminal with a `🐛` pointer.
- 🚦 **Built-in QA & Linter:** Catch dangerous operations before they hit production. The static Linter detects anti-patterns (like accidental `DROP TABLE` or locking queries), while the `dry-run` engine dynamically tests your queries without saving changes.
- 🌱 **Dedicated Stateless Seeder:** Keep your schema definitions (Migrations) strictly separated from your dummy data (Seeders). Comes with idempotent templates to prevent primary-key collision errors.

---

## 📸 Preview

### Beautiful Error DX

![alt text](https://s3.my-playground.space/public-storage/visual-error-highlighting.png)

### Linter Output

![alt text](https://s3.my-playground.space/public-storage/linter-output.png)

---

## 📦 Installation

Ensure you have Go installed on your machine. You can install `go-tecton` directly via `go install`:

    go install github.com/fresmaa/go-tecton/cmd/tecton@latest

---

## 🛠️ Quick Start Guide

### 1. Managing Migrations

Generate a new migration pair (`.up.sql` and `.down.sql`):

    tecton create create_users_table --dir migrations

Apply pending migrations to the database:

    tecton up --db "postgres://user:pass@localhost:5432/dbname?sslmode=disable" --dir migrations

Rollback the last applied migration:

    tecton down --db "postgres://user:pass@localhost:5432/dbname?sslmode=disable" --dir migrations

### 2. Quality Assurance (CI/CD Ready)

Run the static analysis Linter to catch dangerous queries:

    tecton lint --dir migrations

Simulate pending migrations dynamically to ensure syntax validity without committing changes:

    tecton dry-run --db "postgres://user:pass@localhost:5432/dbname?sslmode=disable" --dir migrations

### 3. Data Seeding

Generate a new idempotent seeder file:

    tecton create-seed insert_dummy_admin --dir seeders

Execute all seeder files to populate your database:

    tecton seed --db "postgres://user:pass@localhost:5432/dbname?sslmode=disable" --dir seeders

---

## 🏗️ Architecture & Stack

- **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
- **Terminal UI:** [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Database Driver:** PostgreSQL (extensible driver interface)

---

## 🤝 Support & Feedback

If you encounter any bugs, have feature requests, or need help setting up the CLI, please feel free to reach out or contribute!

- **Bug Reports & Feature Requests:** Please open an issue on our 📁 [GitHub Issues](https://github.com/fresmaa/go-tecton/issues) page.
- **Contributions:** Pull Requests are welcome! Feel free to fork the repo and submit your improvements.

---

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
