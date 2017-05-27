# shareos

shareos is object storage

## start server

```
go run main.go

```


### create bucket

```
 curl -X "PUT"  127.0.0.1:9000/vxt9

```

### list bucket

```
 curl 127.0.0.1:9000/

```
### list object

```
 curl http://127.0.0.1:9000/vxt9

```


### create object 
```
 curl -X "PUT" -T "aa.png"  127.0.0.1:9000/vxt9/hvbt

```

### head object

```
curl -X "HEAD" -v http://127.0.0.1:9000/vxt9/hghk

```

### get object

```
curl -O http://127.0.0.1:9000/vxt9/hghk

```


