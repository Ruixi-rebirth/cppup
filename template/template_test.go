package template

import (
	"strings"
	"testing"
)

func assertContains(t *testing.T, out string, fragments ...string) {
	t.Helper()
	for _, f := range fragments {
		if !strings.Contains(out, f) {
			t.Errorf("missing %q", f)
		}
	}
}

func assertAbsent(t *testing.T, out, fragment string) {
	t.Helper()
	if strings.Contains(out, fragment) {
		t.Errorf("should not contain %q", fragment)
	}
}

func TestCMakeListsExe(t *testing.T) {
	bs, _ := BuildSystemByName("cmake")
	out, _ := CMakeListsExe("myapp", "1.0.0", "17", bs.MinVersion, false)
	assertContains(t, out,
		"cmake_minimum_required(VERSION 3.20)",
		"project(myapp VERSION 1.0.0)",
		"CMAKE_CXX_STANDARD 17",
		"add_executable(${PROJECT_NAME} src/main.cpp)",
		"target_include_directories(${PROJECT_NAME} PRIVATE include)",
	)
	assertAbsent(t, out, "enable_testing")
}

func TestCMakeListsExe_withTests(t *testing.T) {
	bs, _ := BuildSystemByName("cmake")
	out, _ := CMakeListsExe("myapp", "1.0.0", "17", bs.MinVersion, true)
	assertContains(t, out, "enable_testing()", "add_subdirectory(tests)")
}

func TestCMakeListsExe_std(t *testing.T) {
	bs, _ := BuildSystemByName("cmake")
	for _, std := range []string{"11", "14", "20", "23"} {
		out, _ := CMakeListsExe("p", "0.1.0", std, bs.MinVersion, false)
		assertContains(t, out, "CMAKE_CXX_STANDARD "+std)
	}
}

func TestCMakeListsLib_static(t *testing.T) {
	bs, _ := BuildSystemByName("cmake")
	out, _ := CMakeListsLib("mylib", "1.0.0", "20", "STATIC", bs.MinVersion, false)
	assertContains(t, out,
		"CMAKE_CXX_STANDARD 20",
		"add_library(${PROJECT_NAME} STATIC src/${PROJECT_NAME}.cpp)",
		"install(TARGETS ${PROJECT_NAME}",
		"install(DIRECTORY include/",
	)
}

func TestCMakeListsLib_shared(t *testing.T) {
	bs, _ := BuildSystemByName("cmake")
	out, _ := CMakeListsLib("mylib", "1.0.0", "17", "SHARED", bs.MinVersion, false)
	assertContains(t, out, "add_library(${PROJECT_NAME} SHARED src/${PROJECT_NAME}.cpp)")
}

func TestCMakeListsLib_withTests(t *testing.T) {
	bs, _ := BuildSystemByName("cmake")
	out, _ := CMakeListsLib("mylib", "1.0.0", "17", "STATIC", bs.MinVersion, true)
	assertContains(t, out, "enable_testing()", "add_subdirectory(tests)")
}

func TestCMakeListsHeaderOnly(t *testing.T) {
	bs, _ := BuildSystemByName("cmake")
	out, _ := CMakeListsHeaderOnly("mylib", "1.0.0", bs.MinVersion, false)
	assertContains(t, out,
		"add_library(${PROJECT_NAME} INTERFACE)",
		"INTERFACE",
		"install(TARGETS ${PROJECT_NAME})",
		"install(DIRECTORY include/",
	)
}

func TestCMakeListsHeaderOnly_withTests(t *testing.T) {
	bs, _ := BuildSystemByName("cmake")
	out, _ := CMakeListsHeaderOnly("mylib", "1.0.0", bs.MinVersion, true)
	assertContains(t, out, "enable_testing()", "add_subdirectory(tests)")
}

func TestMesonBuildExe(t *testing.T) {
	bs, _ := BuildSystemByName("meson")
	out, _ := MesonBuildExe("myapp", "1.0.0", "17", bs.MinVersion, false)
	assertContains(t, out,
		"'myapp'",
		"'cpp'",
		"version : '1.0.0'",
		"'cpp_std=c++17'",
		"executable('myapp'",
		"'src/main.cpp'",
	)
	assertAbsent(t, out, "subdir(")
}

func TestMesonBuildExe_withTests(t *testing.T) {
	bs, _ := BuildSystemByName("meson")
	out, _ := MesonBuildExe("myapp", "1.0.0", "17", bs.MinVersion, true)
	assertContains(t, out, "subdir('tests')")
}

func TestMesonBuildLib_static(t *testing.T) {
	bs, _ := BuildSystemByName("meson")
	out, _ := MesonBuildLib("mylib", "1.0.0", "20", "static", bs.MinVersion, false)
	assertContains(t, out,
		"static_library('mylib'",
		"'src/mylib.cpp'",
		"mylib_dep = declare_dependency(",
		"link_with : mylib_lib",
	)
}

func TestMesonBuildLib_shared(t *testing.T) {
	bs, _ := BuildSystemByName("meson")
	out, _ := MesonBuildLib("mylib", "1.0.0", "17", "shared", bs.MinVersion, false)
	assertContains(t, out, "shared_library('mylib'")
}

func TestMesonBuildHeaderOnly(t *testing.T) {
	bs, _ := BuildSystemByName("meson")
	out, _ := MesonBuildHeaderOnly("mylib", "1.0.0", "23", bs.MinVersion, false)
	assertContains(t, out,
		"'mylib'",
		"'cpp'",
		"'cpp_std=c++23'",
		"mylib_dep = declare_dependency(",
	)
}

func TestMesonStdOpt(t *testing.T) {
	bs, _ := BuildSystemByName("meson")
	for _, std := range []string{"11", "14", "17", "20", "23"} {
		out, _ := MesonBuildExe("p", "0.1.0", std, bs.MinVersion, false)
		assertContains(t, out, "'cpp_std=c++"+std+"'")
	}
}

func TestTestCMakeLists_googletest(t *testing.T) {
	fw, _ := TFrameworkByName("googletest")
	out, _ := TestCMakeLists("myapp", "exe", "googletest", fw.DefaultTag)
	assertContains(t, out,
		fw.DefaultTag,
		"GTest::gtest",
		"gtest_discover_tests(tests)",
	)
	assertAbsent(t, out, "GTest::gtest_main")
}

func TestTestCMakeLists_catch2(t *testing.T) {
	fw, _ := TFrameworkByName("catch2")
	out, _ := TestCMakeLists("myapp", "exe", "catch2", fw.DefaultTag)
	assertContains(t, out,
		fw.DefaultTag,
		"Catch2::Catch2WithMain",
		"catch_discover_tests(tests)",
		"CMAKE_MODULE_PATH ${catch2_SOURCE_DIR}/extras",
	)
}

func TestTestCMakeLists_doctest(t *testing.T) {
	fw, _ := TFrameworkByName("doctest")
	out, _ := TestCMakeLists("myapp", "exe", "doctest", fw.DefaultTag)
	assertContains(t, out,
		fw.DefaultTag,
		"doctest::doctest",
		"doctest_discover_tests(tests)",
		"CMAKE_MODULE_PATH ${doctest_SOURCE_DIR}/scripts/cmake",
	)
}

func TestTestCMakeLists_libLinksTarget(t *testing.T) {
	fw, _ := TFrameworkByName("catch2")
	// lib project should add the library to the link line
	out, _ := TestCMakeLists("mylib", "lib-static", "catch2", fw.DefaultTag)
	assertContains(t, out, "mylib")

	// exe project should not add an exe target to the link line
	outExe, _ := TestCMakeLists("myapp", "exe", "catch2", fw.DefaultTag)
	if strings.Contains(outExe, "Catch2::Catch2WithMain myapp") {
		t.Error("exe project should not link against exe target")
	}
}

func TestTestMesonBuild_googletest(t *testing.T) {
	out, _ := TestMesonBuild("myapp", "exe", "googletest")
	assertContains(t, out, "dependency('gtest'")
}

func TestTestMesonBuild_catch2(t *testing.T) {
	out, _ := TestMesonBuild("myapp", "exe", "catch2")
	assertContains(t, out, "dependency('catch2-with-main'")
}

func TestTestMesonBuild_doctest(t *testing.T) {
	out, _ := TestMesonBuild("myapp", "exe", "doctest")
	assertContains(t, out, "dependency('doctest'")
}

func TestTestMesonBuild_libDep(t *testing.T) {
	out, _ := TestMesonBuild("mylib", "lib-static", "catch2")
	assertContains(t, out, "mylib_dep")

	outExe, _ := TestMesonBuild("myapp", "exe", "catch2")
	assertAbsent(t, outExe, "myapp_dep")
}

func TestTestSourceCpp_googletest(t *testing.T) {
	out, _ := TestSourceCpp("googletest")
	assertContains(t, out, "#include <gtest/gtest.h>", "RUN_ALL_TESTS()")
}

func TestTestSourceCpp_catch2(t *testing.T) {
	out, _ := TestSourceCpp("catch2")
	assertContains(t, out, "#include <catch2/catch_test_macros.hpp>", "TEST_CASE")
}

func TestTestSourceCpp_doctest(t *testing.T) {
	out, _ := TestSourceCpp("doctest")
	assertContains(t, out, "DOCTEST_CONFIG_IMPLEMENT_WITH_MAIN", "#include <doctest/doctest.h>")
}

func TestMesonWrap_versions(t *testing.T) {
	for _, fw := range TFrameworks {
		out, _ := MesonWrap(fw.Name, fw.DefaultTag)
		if !strings.Contains(out, fw.DefaultTag) {
			t.Errorf("MesonWrap(%q) missing version %q", fw.Name, fw.DefaultTag)
		}
	}
}

func TestMesonWrapFilename(t *testing.T) {
	cases := map[string]string{
		"googletest": "gtest.wrap",
		"catch2":     "catch2.wrap",
		"doctest":    "doctest.wrap",
	}
	for fw, want := range cases {
		if got := MesonWrapFilename(fw); got != want {
			t.Errorf("MesonWrapFilename(%q) = %q, want %q", fw, got, want)
		}
	}
}

func TestMainCpp(t *testing.T) {
	out, _ := MainCpp()
	assertContains(t, out, "#include <iostream>", "int main()", "Hello, World!")
}

func TestLibHpp(t *testing.T) {
	out, _ := LibHpp("mylib")
	assertContains(t, out, "#pragma once", "namespace mylib", "} // namespace mylib")
}

func TestLibCpp(t *testing.T) {
	out, _ := LibCpp("mylib")
	assertContains(t, out, "#include <mylib/mylib.hpp>", "namespace mylib")
}

func TestHeaderOnlyHpp(t *testing.T) {
	out, _ := HeaderOnlyHpp("mylib")
	assertContains(t, out, "#pragma once", "namespace mylib")
}
