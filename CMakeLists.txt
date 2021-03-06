CMAKE_MINIMUM_REQUIRED (VERSION 2.8)
PROJECT (indexing)

INCLUDE (FindCouchbaseGo)
INCLUDE (FindCouchbaseJemalloc)

get_filename_component(JEMALLOC_LIB_DIR ${JEMALLOC_LIBRARIES} DIRECTORY)

SET (_ldflags)
IF (APPLE)
  # On OS X 10.11 (El Capitan) upwards we can no longer use DYLD_LIBRARY_PATH to locate
  # runtime dependancies. Instead add the location of forestdb etc to the actual
  # binary itself.
  SET (_ldflags "-extldflags '-Wl,-rpath,@executable_path/../lib'")
ENDIF ()

SET (_indexer_ldflags "${_ldflags}")
IF (WIN32)
  INCLUDE (PlatformIntrospection)
  _DETERMINE_ARCH (_arch)
  IF (NOT _arch STREQUAL x86)
    SET (_indexer_ldflags "${_indexer_ldflags} -linkmode internal")
  ENDIF ()
ENDIF ()

GET_FILENAME_COMPONENT (CURL_LIBRARY_DIR "${CURL_LIBRARIES}" DIRECTORY)

SET(CGO_INCLUDE_DIRS "${FORESTDB_INCLUDE_DIR};${Platform_SOURCE_DIR}/include;${CURL_INCLUDE_DIR}")
SET(CGO_LIBRARY_DIRS "${FORESTDB_LIBRARY_DIR};${Platform_BINARY_DIR};${CURL_LIBRARY_DIR}")

IF(JEMALLOC_FOUND)
    SET (ENV{CGO_CFLAGS} "$ENV{CGO_CFLAGS} -DJEMALLOC=1")
    SET (CGO_INCLUDE_DIRS "${CGO_INCLUDE_DIRS};${JEMALLOC_INCLUDE_DIR}")
    SET (CGO_LIBRARY_DIRS "${CGO_LIBRARY_DIRS};${JEMALLOC_LIB_DIR}")
    SET (TAGS "jemalloc")
ENDIF(JEMALLOC_FOUND)

GoInstall (TARGET indexer PACKAGE github.com/couchbase/indexing/secondary/cmd/indexer
  GOPATH "${PROJECT_SOURCE_DIR}/../../../.." "${GODEPSDIR}"
  DEPENDS forestdb
  CGO_INCLUDE_DIRS "${CGO_INCLUDE_DIRS}"
  CGO_LIBRARY_DIRS "${CGO_LIBRARY_DIRS}"
  GOTAGS "${TAGS}"
  LDFLAGS "${_indexer_ldflags}"
  INSTALL_PATH bin OUTPUT indexer
  GOVERSION 1.7.3)

GoInstall (TARGET projector PACKAGE github.com/couchbase/indexing/secondary/cmd/projector
  GOPATH "${PROJECT_SOURCE_DIR}/../../../.." "${GODEPSDIR}"
  CGO_INCLUDE_DIRS "${CGO_INCLUDE_DIRS}"
  CGO_LIBRARY_DIRS "${CGO_LIBRARY_DIRS}"
  GOTAGS "${TAGS}"
  INSTALL_PATH bin
  GOVERSION 1.7.3)

GoInstall (TARGET cbindex PACKAGE github.com/couchbase/indexing/secondary/cmd/cbindex
  GOPATH "${PROJECT_SOURCE_DIR}/../../../.." "${GODEPSDIR}"
  INSTALL_PATH bin
  GOVERSION 1.7.3)

GoInstall (TARGET cbindexperf PACKAGE github.com/couchbase/indexing/secondary/cmd/cbindexperf
  GOPATH "${PROJECT_SOURCE_DIR}/../../../.." "${GODEPSDIR}"
  INSTALL_PATH bin
  GOVERSION 1.7.3)

GoInstall (TARGET cbindexplan PACKAGE github.com/couchbase/indexing/secondary/cmd/cbindexplan
  GOPATH "${PROJECT_SOURCE_DIR}/../../../.." "${GODEPSDIR}"
  DEPENDS forestdb
  CGO_INCLUDE_DIRS "${CGO_INCLUDE_DIRS}"
  CGO_LIBRARY_DIRS "${CGO_LIBRARY_DIRS}"
  GOTAGS "${TAGS}"
  LDFLAGS "${_indexer_ldflags}"
  INSTALL_PATH bin OUTPUT cbindexplan
  GOVERSION 1.7.3)

GoInstall(TARGET plasma_dump PACKAGE github.com/couchbase/nitro/cmd/plasma_dump
  GOPATH "${CMAKE_SOURCE_DIR}/goproj" "${GODEPSDIR}"
  INSTALL_PATH bin OUTPUT plasma_dump
  GOVERSION 1.7.3)
