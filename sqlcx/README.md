# SQLC eXtended

This is a package designed to provide functionality for use when using sqlc in a Go app.


The directory structure intended for this is: 

```
./sqlc.yaml
./schema.sql
./queries.sql
./sqlcx/ddl.go:         EMBEDS ../schema.sql
./sqlcx/schema.sql:     SYMLINKS ../schema.sql 
```

Note the **SYMLINK**; if it is not there you need to add it.  

The SYMLINK is required because of an annoying limitation of Go that an embedded file must be in the same directory or a child directory of the Go file that embeds it. 

