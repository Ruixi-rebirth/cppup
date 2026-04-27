cmake_minimum_required(VERSION {{.MinVersion}})
project({{.ProjectName}} VERSION {{.Version}})

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

add_library(${PROJECT_NAME} INTERFACE)

target_include_directories(${PROJECT_NAME}
  INTERFACE
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>
)

install(TARGETS ${PROJECT_NAME})
install(DIRECTORY include/ DESTINATION include)
{{- if .WithTests}}

enable_testing()
add_subdirectory(tests)
{{- end}}
