# go-reflector

Go reflector is a library that makes working with reflection in Go easier and safer.
It provides a wrappers for easy acces to struct fields, safe setting of values and more.

One principal of the library is to never panic, but return nil values or errors instead, 
unlike the reflect package of the standard library.

## Install

```bash
go get github.com/theduke/go-reflector
```

## Examples

### Working with structs.

```go
import "github.com/theduke/go-reflector"


```

## Additional information

### Changelog

[Changelog](https://github.com/theduke/go-reflector/blob/master/CHANGELOG.md)

### Versioning

This project follows [SemVer](http://semver.org/).

### License

This project is under the [MIT license](https://opensource.org/licenses/MIT).
