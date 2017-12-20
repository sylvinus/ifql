# IFQL Language

This document details the design of the IFQL langauage.
If you are looking for usage information on the langauage see the README.md.

# Overview

The IFQL langauage is used to construct query specifications.

# Syntax

The langauage syntax is defined by the ifql/ifql.peg grammar.

## Keyword Arguments

IFQL uses keyword arguments for ALL arguments to functions.
Keyword arguments enable iterative improvements to the langauage while remaining backwards compatible.

### Default Arguments

Since all arguments are keyword arguments and there are no positional arguments its possible for any argument to have a default value.
If an argument is not specified at call time, then if the argument has a default it is used, otherwise an error occurs.

## Abstract Syntax Tree

The abstract syntax tree (AST) of IFQL is closely modeled after the javascript AST.
Using the javascript AST provides a good foundation for organization and structure of the syntax tree.
Since IFQL is so similar to javascript this design works well.

# Semantics

Currently IFQL does not have a AST to semantic graph transformation, all components simply consume the AST directly.
Using a semantic graph representation of the IFQL, will enable highlevel meaning to be specified programatically.

For example since IFQL uses the javascript AST structures, arguments to a function are represented as a single positional argument that is always an object expression.
A semantic graph can validate that the AST correctly follows these semantics, and use structures more suited IFQLs semantic abstractions instead of a raw AST.

The semantic structures are to be designed to facilitate the interpretation and compilation of IFQL.

# Interpretation

IFQL is primarily an interpreted language.
The implementation of the IFQL interpreter can be found in the `ifql/eval.go` file.


# Compilation and Go Runtime

A subset of IFQL can be compiled into a runtime hosted in Go.
The subset consists of only pure functions.
Meaning a function defintion in IFQL can be compiled and then called repeatedly with different arguments.
The function must be pure meaning it has no side effects.
Other language feature like imports etc are not supported.

This runtime is entirely not portable.
The runtime consists of Go types that have been constructed based on the IFQL function being compiled.
Those types are not serializable and cannot be transported to other systems or environments.
This design is intended to limit the scope under which compilation must be supported.

# Features

This sections details various features of the language.

## Functions

IFQL supports defining functions.

Example:

```
add = (a,b) => a + b

add(a:1, b:2) // 3
```

Functions can be assigned to identifiers and can call other functions.
Functions are first class types within IFQL.

## Scoping

IFQL uses lexical scoping.
Scoping boundaries occur at functions.

Example:

```
x = 5
addX = (a) => a + x

add(a:1) // 6
```

The `x` referred to in the `addX` function is the same as is defined in the toplevel scope.

Scope names can be changed for more specific scopes.

Example:

```
x = 5

add = (x,y) => x + y

add(x:1,y:2) // 3
```

In this example the `x = 5` definition is unused, as the `add` function defines it own local identifier `x` as a parameter.

## Imports

IFQL supports importing packages.
A package is a collection of IFQL scripts, typically represented as `.ifql` files in a directory.

Importing a package evaluates the contents of the package and makes the package scope available under the package name in the current scope.

Example:


```
// file: foo/bar.ifql
package foo

fortytwo = 42
```

```
// file query.ifql
import "foo"

foo.fortytwo / 6 == 7 // true
```



### Versioning

When importing a package a version may be specified.
If not then the latest version is used.

Example:

```
import "foo" =1.0.0
```

Import the `1.0.0` version of the `foo` package.

Additionally semantic version operators may be sepecified.

Example:

```
import "foo" ~1.0.0 // Allow patch versions i.e. 1.0.x
import "foo" ^1.2.0 // Allow minor versions i.e. 1.x.x
import "foo" // Use latest version
```

Version strings may have a leading `v` which will be ignored.

### Registry

Eventually we will want a registry to host published packages.
In the meantime allowing popular VCS hosting sites to act as defacto registries will enable easy sharing of IFQL packages.

If a package import path begins with a name that is recognized then the package can be downloaded locally.

Example:

```
import "github.com/influxdata/ifql-pkg/math"

math.pow(x:5.0,y:3.0) // 5^3 = 125
```

Running the command `ifql install` will search the current directory for `.ifql` source files and download any imports specified within.

If installing a specific version, the version string is assumed to be a valid VCS referrence.
