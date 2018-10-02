<h1 align="center">
    <br>
    <img width="400" src="https://github.com/solodynamo/poolboy/blob/master/poolboy.png" alt="gopool">
    <br>
    <br>
    <br>
</h1>

- To run benchmark := go test -test.bench=<benchmarkfnname>
- To run tests := go test -v

- To do : Replace doubly linked list implementation with slices
BenchmarkSlices-4       100000000               14.3 ns/op
BenchmarkLists-4         5000000               275 ns/op