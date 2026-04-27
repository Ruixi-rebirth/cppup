include(FetchContent)
FetchContent_Declare(
  Catch2
  URL                      {{.ArchiveURL}}
  DOWNLOAD_EXTRACT_TIMESTAMP TRUE
)
FetchContent_MakeAvailable(Catch2)
list(APPEND CMAKE_MODULE_PATH ${catch2_SOURCE_DIR}/extras)

add_executable(tests test_main.cpp)
target_link_libraries(tests PRIVATE Catch2::Catch2WithMain{{if .LibName}} {{.LibName}}{{end}})

include(CTest)
include(Catch)
catch_discover_tests(tests)
