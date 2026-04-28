cmake_minimum_required(VERSION {{.MinVersion}})
project({{.ProjectName}} VERSION {{.Version}})

set(CMAKE_CXX_STANDARD {{.CxxStd}})
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

#[[ Dependencies: use find_package() for system libs or FetchContent for others.
Example with FetchContent:
  include(FetchContent)
  FetchContent_Declare(fmt URL https://github.com/fmtlib/fmt/archive/refs/tags/11.0.2.tar.gz DOWNLOAD_EXTRACT_TIMESTAMP TRUE)
  FetchContent_MakeAvailable(fmt) ]]

add_executable(${PROJECT_NAME} src/main.cpp)

target_include_directories(${PROJECT_NAME} PRIVATE include)

target_compile_options(${PROJECT_NAME} PRIVATE -Wall -Wextra -Wpedantic)

# target_link_libraries(${PROJECT_NAME} PRIVATE fmt::fmt)

# Installs to CMAKE_INSTALL_PREFIX/bin (default: /usr/local, or $out when using Nix)
install(TARGETS ${PROJECT_NAME})
{{- if .WithTests}}

enable_testing()
add_subdirectory(tests)
{{- end}}
