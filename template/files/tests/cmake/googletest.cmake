include(FetchContent)
FetchContent_Declare(
  googletest
  URL                      {{.ArchiveURL}}
  DOWNLOAD_EXTRACT_TIMESTAMP TRUE
)
set(gtest_force_shared_crt ON CACHE BOOL "" FORCE)
FetchContent_MakeAvailable(googletest)

add_executable(tests test_main.cpp)
target_link_libraries(tests PRIVATE GTest::gtest{{if .LibName}} {{.LibName}}{{end}})

include(GoogleTest)
gtest_discover_tests(tests)
