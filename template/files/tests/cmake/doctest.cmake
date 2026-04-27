include(FetchContent)
FetchContent_Declare(
  doctest
  URL                      {{.ArchiveURL}}
  DOWNLOAD_EXTRACT_TIMESTAMP TRUE
)
FetchContent_MakeAvailable(doctest)
list(APPEND CMAKE_MODULE_PATH ${doctest_SOURCE_DIR}/scripts/cmake)

add_executable(tests test_main.cpp)
target_link_libraries(tests PRIVATE doctest::doctest{{if .LibName}} {{.LibName}}{{end}})

include(doctest)
doctest_discover_tests(tests)
