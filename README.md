# GO LIB
[![](https://travis-ci.com/Tracking-SYS/go-lib.svg?branch=main)](https://travis-ci.com/github/Tracking-SYS/go-lib/builds)
## Preparation

1. Install **golangci-lint** for Go linters aggregator. Follow [here](https://golangci-lint.run/usage/install/#local-installation)

## How to use

1. This library use under as private package. So remember set env ```GOPRIVATE``` before import to it into your project

        export GOPRIVATE=github.com/Tracking-SYS

2. Download package with latest version
        
        go get -u github.com/Tracking-SYS/go-lib

3. Import specific package

        import (
                "github.com/Tracking-SYS/go-lib/http"
                "github.com/Tracking-SYS/go-lib/trace"

                 ....
        )


***NOTE: This lib still be under development phase. Maybe not compatible***

_Update to: 2021-05-13_
