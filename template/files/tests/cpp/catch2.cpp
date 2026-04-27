#include <catch2/catch_test_macros.hpp>

TEST_CASE("basic assertions", "[basic]") {
    REQUIRE(1 + 1 == 2);
    CHECK(true);
}
