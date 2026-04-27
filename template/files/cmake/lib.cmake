cmake_minimum_required(VERSION {{.MinVersion}})
project({{.ProjectName}} VERSION {{.Version}})

set(CMAKE_CXX_STANDARD {{.CxxStd}})
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Dependencies: use find_package() for system libs or FetchContent for others.

add_library(${PROJECT_NAME} {{.LibType}} src/${PROJECT_NAME}.cpp)

target_include_directories(${PROJECT_NAME}
  PUBLIC
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>
  PRIVATE
    src
)

target_compile_options(${PROJECT_NAME} PRIVATE -Wall -Wextra -Wpedantic)

install(TARGETS ${PROJECT_NAME}
  ARCHIVE DESTINATION lib
  LIBRARY DESTINATION lib
  RUNTIME DESTINATION bin
)
install(DIRECTORY include/ DESTINATION include)
{{- if .WithTests}}

enable_testing()
add_subdirectory(tests)
{{- end}}
