# cppup

A scaffold tool for C++ projects, inspired by `cargo`.
Supports CMake and Meson, with deep Nix integration, test frameworks, and modern C++ development workflows.

## Features

- **Blazing Fast Scaffolding**: Create professional C++ projects in seconds with standard directory structures.
- **Smart Interactivity**: Interactive mode supports **first-letter matching** (e.g., just type `e` for `exe`).
- **Nix-First**: Automated generation of `flake.nix`, `treefmt.nix`, and `.envrc` for reproducible environments.
- **LSP Ready**: Automatically manages `compile_commands.json` symlinks for `clangd` support.
- **Sanitizer Support**: One-flag enablement for AddressSanitizer and UndefinedBehaviorSanitizer (`--asan`).
- **Lifecycle Management**: Build, run, test, format, and check projects from any sub-directory.
- **Persistent Metadata**: Tracks project state in `.cppup` for intelligent incremental updates.

## Install

```bash
# Nix
nix profile add github:Ruixi-rebirth/cppup

# Go
go install github.com/Ruixi-rebirth/cppup@latest
```

## Commands

### `cppup new [name]`

Create a new project in a new subdirectory. Interactive when stdin is a TTY (with smart letter matching), non-interactive otherwise.

```bash
cppup new myapp
cppup new mylib --type lib-static --std 20 --build meson
cppup new myapp --tests catch2 --clang-format --nix
```

### `cppup init [name]`

Scaffolds a project into the current directory. Name defaults to the current directory's basename.

### Flags & Options

| Flag              | Default | Description                                     |
| ----------------- | :-----: | ----------------------------------------------- |
| `--type`          |  `exe`  | `exe`, `lib-static`, `lib-shared`, `lib-header` |
| `--build`         | `cmake` | `cmake` or `meson`                              |
| `--build-version` |    -    | Minimum build system version (e.g. `3.25`)      |
| `--std`           |  `17`   | C++ standard: `11`, `14`, `17`, `20`, `23`      |
| `--version`       | `0.1.0` | Project version                                 |
| `--tests`         |    -    | `googletest`, `catch2`, or `doctest`            |
| `--tests-version` |    -    | Test framework git tag (e.g. `v3.7.1`)          |
| `--clang-format`  |    -    | Generate `.clang-format`                        |
| `--clang-tidy`    |    -    | Generate `.clang-tidy`                          |
| `--nix`           |    -    | Generate `flake.nix` and `.envrc`               |
| `--no-git`        |    -    | Skip `git init`                                 |

### `cppup build`

```bash
cppup build            # debug   → build/debug/
cppup build --release  # release → build/release/
cppup build --asan     # asan    → build/asan/
```

- Prefers Ninja when available.
- Debug and ASan builds symlink `compile_commands.json` to project root for LSP.
- `--asan` enables `-fsanitize=address,undefined`.

### `cppup run`

Build and run. Only available for `exe` projects. Supports arguments passing:
`cppup run --release -- arg1 arg2`

### `cppup test`

Build and run tests using `ctest` or `meson test`.

### `cppup add`

Add components or dependencies to an existing project. Smartly reads metadata from `.cppup`.

```bash
cppup add tests --framework catch2
cppup add clang-format
cppup add clang-tidy
cppup add nix

# Dependency Management
cppup add dep fmt              # Add from WrapDB (Meson only)
cppup add dep --git https://github.com/fmtlib/fmt.git --tag v11.0.2
cppup add dep --url https://example.com/lib.tar.gz --name mylib
```

### `cppup deps`

List all configured dependencies in the project.

### `cppup remove <name>`

Remove a dependency or component.

```bash
cppup remove dep fmt
```

### `cppup fmt`

Format all source files.

- **With Nix**: Runs `nix fmt` (using treefmt-nix).
- **Standalone**: Runs `clang-format -i` on all C++ source/header files.

### `cppup check`

Run `clang-tidy` on source files.

- `cppup check --no-tests`: Skip scanning the `tests/` directory.

### `cppup install`

Build (release) and install.

```bash
cppup install              # → /usr/local
cppup install ~/.local     # → ~/.local
sudo cppup install         # → /usr/local (if not writable)
```

### `cppup clean`

Remove `build/` artifacts and symlinks.

## Project Structure

Example for `exe` + `cmake` + `catch2`:

```
myapp/
├── .cppup              # Metadata (name, version, std...)
├── CMakeLists.txt      # Root build file
├── src/
│   └── main.cpp
├── include/
│   └── myapp/          # Header directory
├── tests/
│   ├── CMakeLists.txt
│   └── test_main.cpp
├── .gitignore
└── flake.nix           # Optional Nix integration
```
