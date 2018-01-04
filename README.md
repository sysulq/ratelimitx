# ratelimitx
A simple ratelimit for golang, implemented with memcache and gobreak, aims on high availability.

test results
===
```
PASS
coverage: 85.2% of statements
ok  	ratelimitx	1.811s
Success: Tests passed.
```

benchmark results
===
```
$ go test -bench=.
goos: windows
goarch: amd64
pkg: ratelimitx
BenchmarkAllowSecond-4                                      5000            253200 ns/op
BenchmarkAllowMinute-4                                      5000            237330 ns/op
BenchmarkAllowHour-4                                       10000            247800 ns/op
BenchmarkMemcacheUnavailableWithFallback-4               1000000              1075 ns/op
BenchmarkMemcacheUnavailableWithoutFallback-4            2000000               918 ns/op
PASS
ok      go_third_party/app/lib/ratelimitx       12.950s
```
