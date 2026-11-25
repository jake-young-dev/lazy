# lazy

Lazy is a helper tool when working with [sqlx](https://github.com/jmoiron/sqlx) and [sqlmock](https://github.com/DATA-DOG/go-sqlmock) that generates mock data for unit tests based on the example struct.

<hr />

- [RandomGenerate](#randomgenerate)
    * [Config](#config)
    * [Mocks](#mocks)
    * [Tags](#tags)
    * [Example](#example)

<hr />

## RandomGenerate
Generates random data for basic types and slices to be used with [sqlmock](https://github.com/DATA-DOG/go-sqlmock). Slices are given an arbitrary amount entries See [Mocks](#mocks) for more information. Lazy supports many bind types using sqlx's Rebind method, however, due to this conversion queries that use operators like postgres' jsonb "?" operator cannot be paired with the BindQuestion option as sqlx will greedily replace all "?" causing syntax errors
```go
func RandomGenerate(m Config) (*Mock, error)
```

<hr />

### Config
```go
type Config struct {
	Query    string
	Example  any
	Keys     []any
	RowCount int
}
```
The Config struct is used to generate the data:
- <b>Query</b> the sql query that is being run in the function being tested
- <b>Example</b> represents the struct that will be used in the sql scan, this struct is used to detect which fields are expected in the query. Fields must have "db" tag to be parsed
- <b>Keys</b> primary keys to be hardcoded in mock rows, if needed. Primary keys are detected by tags in the example struct, see [Tags](#tags) for more information
- <b>RowCount</b> the amount of mock rows to be generated, if you are setting primary keys this number must be equal to the amount of keys supplied

<hr />

### Mocks
```go
type Mock struct {
	Query   string
	Columns []string
	Rows    [][]driver.Value
}
```
The Mock struct is returned by RandomGenerate:
- <b>Query</b> represents the query string expected by [ExpectQuery](https://pkg.go.dev/github.com/DATA-DOG/go-sqlmock#SqlmockCommon.ExpectQuery). The original query is rebound using the bindvar type [sqlx.AT](https://pkg.go.dev/github.com/jmoiron/sqlx#AT)
- <b>Columns</b> slice of expected column names for [sqlmock.NewRows](https://pkg.go.dev/github.com/DATA-DOG/go-sqlmock#NewRows)
- <b>Rows</b> represents the row data for [AddRows](https://pkg.go.dev/github.com/DATA-DOG/go-sqlmock#Rows.AddRows)

<hr />

### Tags
Struct tags are required on the example struct to generate the data
- <b>lazy</b> optional, set this tag to <b>key</b> to mark the primary key field on the example struct, if no Keys are provided in the Config struct this tag will be ignored
- <b>db</b> required for mock columns to be generated, columns will be based on tag value

<hr />

### Example
This program is a psuedo-code example of the mock generation. It takes the SqlResultsExample and uses it to generate two rows of mock data using the Keys (1, 6) as the Key, signified by the lazy tag. See [Example output](#example-output-rows) for the generated sql rows.
```go
func ExampleTest() {
    ...

    type SqlResultsExample struct {
        Key     string  `db:"key" lazy:"key"`
        Count   int     `db:"count"`
    }
    cfg := lazy.Config{
	    Query:      "SELECT key, count FROM table WHERE key IN (?, ?)",
	    Example:    SqlResultsExample{},
	    Keys:       []any{ 1, 6 },
	    RowCount:   2, 
        BindType: lazy.BindQuestion, 
    }
    mock, err := lazy.GenerateRandom(cfg)
    if err != nil {
        panic(err)
    }
    rows := sqlmock.NewRows(mock.Columns).AddRows(mock.Rows...)
    sqlmock.ExpectQuery(mock.Query).WithArgs(1, 6).WillReturnRows(rows)

    ...

}
```
#### Example of mocked rows from ExampleTest
|key|count|
|---|-----|
|1|649018089772805963|
|6|1066940846275557682|
