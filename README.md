# V3IO Frames

[![Build Status](https://travis-ci.org/v3io/frames.svg?branch=master)](https://travis-ci.org/v3io/frames)
[![GoDoc](https://godoc.org/github.com/v3io/frames?status.svg)](https://godoc.org/github.com/v3io/frames)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

V3IO Frames (**"Frames"**) is a high-speed server and client library for accessing time-series (TSDB), NoSQL, and stream data in the [Iguazio Data Science Platform](https://www.iguazio.com) (**"the platform"**).

#### In This Document

- [API Reference](#api-reference)
- [Contributing](#contributing)
- [LICENSE](#license)

## API Reference

- [Overview](#overview)
- [Client Object Constructor](#client-constructor)
- [Client Operations](#client-operations)

### Overview

Frames currently supports basic CRUD operations &mdash; create, read, update (write), and delete &mdash; for the following backend types:

<a id="backend-types"></a>
- `"kv"` &mdash; a platform NoSQL table (a.k.a "KV table")
- `"stream"` &mdash; a platform data stream
- `"tsdb"` &mdash; a time-series database (TSDB)
- `"csv"` &mdash; a comma-separated-value (CSV) file, used for testing purposes

To use Frames, you need to import the **v3io_frames** library.
For example:
```python
import v3io_frames as v3f
```

Then, you need to create an instance of the client object (see [Client Constructor](#client-constructor)), which you can use to perform different operations on the supported backend types (see [Client Operations](#client-operations)).

> **Note:** The API reference in this document currently uses the syntax of the Python Frames API.
> However, there's also a similar API for Golang (Go).
<!-- SLSL TODO NOWNOWNOW: Edit + edit the rest of the doc to only use Python.
  Tal said that the high-level user API isn't supported in Go. -->

<a id="client-constructor"></a>
### Client Object Constructor

All Frames operations are executed via the `Client` object.

- [Syntax](#client-constructor-syntax)
- [Parameters and Data Members](#client-constructor-parameters)
- [Example](#client-constructor-example)

<a id="client-constructor-syntax"></a>
### Syntax

```python
Client(address, user, password, token, container)
```

<a id="client-constructor-parameters"></a>
#### Parameters and Data Members

<a id="client-object-data-access-auth-note"></a>
> **Data-Access Authentication Note**
> <br/>
> When running the Frames code from the managed Jupyter Notebook service of an Iguazio Data Science Platform cluster, the credentials for authenticating access to the backend data are set implicitly, so you don't need to set any of the authentication parameters &mdash; `user`, `password`, or `token`.
> Otherwise, you need to either use the `user` and `password` constructor parameters to provide a username and password, or use the `token` parameter to provide an access token.

- <a id="client-param-address"></a>**address** &mdash; The `framesdb` backend address.
  This parameter should always be set to `framesd:8081`.

  - **Type:** String
  - **Requirement:** Required 

- <a id="client-param-container"></a>**container** &mdash; The name of the platform data container that contains the backend data.
  For example, `"bigdata"` or `"users"`.

  - **Type:** String
  - **Requirement:** Optional 
  - **Default Value:** `"bigdata"`
  <!-- SLSL PENDING-DEV TODO: Pending answers from Tal. -->

- <a id="client-param-user"></a>**user** &mdash; The username of a platform user with permissions to access the backend data.

  - **Type:** String
  - **Requirement:** See the [data-access authentication note](#client-object-data-access-auth-note).
    <!-- SLSL PENDING-DEV TODO: I'm awaiting confirmation from Tal. -->

- <a id="client-param-password"></a>**password** &mdash; A platform password for the user specified in the `user` parameter.

  - **Type:** String
  - **Requirement:** See the [data-access authentication note](#client-object-data-access-auth-note).
    <!-- SLSL PENDING-DEV TODO: I'm awaiting confirmation from Tal. -->

- <a id="client-param-token"></a>**token** &mdash; A platform access token for accessing the backend data.
  To get this token, select the user profile icon on any platform dashboard page, select **Access Tokens**, and create a new token or copy an existing token.

  - **Type:** String
  - **Requirement:** See the [data-access authentication note](#client-object-data-access-auth-note).

<a id="client-constructor-example"></a>
#### Example

```python
import v3io_frames as v3f
client = v3f.Client("framesd:8081", user="iguazio", password="mypass", container="users")
```
<a id=client-operations""></a>
### Client Operations

After you create a Frames [client object](#client-constructor), you can use the supported object methods to perform different data operations.

- [Common Operation Method Parameters](#client-common-operation-method-params)
- [Create](#create)
- [Write](#write)
- [Read](#read)
- [Delete](#delete)
- [Execute](#execute)

<a id="client-common-operation-method-params"></a>
### Common Operation Method Parameters

All client-object operation methods receive the following common parameters:

- **backend** &mdash; The backend data type for the operation.
  See the backend-types descriptions in the [overview](#backend-types).

  - **Type:** String
  - **Valid Values:** `"csv"` | `"kv"` | `"stream"` | `"tsdb"`
  - **Requirement:** Required

- **table** &mdash; The relative path to the backend data (such as a TSDB or NoSQL table directory or a stream directory) within the target platform data container, as configured for the client object.
  For example, `"mytable"` or `"examples/tsdb/my_metrics"`.

  - **Type:** String
  - **Requirement:** Required

Additional parameters are described for each operation.

<!-- SLSL: I kept the sentence-case spelling to keep it generic for Python and
  Go. TODO: Add info about the Python method names used in the examples
  (lower-case names) and the matching Go names (sentence-case names + "Exec"
  vs. "execute" in Python). NOWNOW -->
#### Create

Creates a new table for the desired backend.
Not all backends require a table to be created prior to ingestion.
For example, a NoSQL table will be created while ingesting new data, on the other hand since TSDB tables have mandatory fields you need to create a table before ingesting new data.

```python
client.create(backend=<backend>, table=<table>, attrs=<backend_specific_attributes>)
```

##### Backend-Specific Parameters

###### TSDB

- rate
- aggregates (optional)
- aggregation-granularity (optional)

For detailed info on these parameters please visit [TSDB](https://github.com/v3io/v3io-tsdb#v3io-tsdb) docs.

Example:
```python
client.create("tsdb", "/mytable", attrs={"rate": "1/m"})
```

###### Stream

- shards=1 (optional)
- retention_hours=24 (optional)

For detailed info on these parameters please visit [Stream](https://www.iguazio.com/docs/concepts/latest-release/streams) docs.

Example:
```python
client.create("stream", "/mystream", attrs={"shards": "6"})
```

#### Write
Writes a DataFrame into one of the supported backends.

Common write parameters:
- dfs - list of DataFrames to write
- index_cols=None (optional) - specify specific index columns, by default DataFrame's index columns will be used.
- labels=None (optional)
- max_in_message=0 (optional)
- partition_keys=None (Not yet supported)

Example:
```python
data = [["tom", 10], ["nick", 15], ["juli", 14]]
df = pd.DataFrame(data, columns = ["name", "age"])
df.set_index("name")
client.write(backend="kv", table="mytable", dfs=df)
```

##### Backend-Specific Parameters

###### NoSQL

- expression=" " (optional) - for detailed information on update expressions see [docs](https://www.iguazio.com/docs/reference/latest-release/expressions/update-expression/)
- condition=" " (optional) - for detailed information on condition expressions see [docs](https://www.iguazio.com/docs/reference/latest-release/expressions/condition-expression/)

Example:
```python
data = [["tom", 10, "TLV"], ["nick", 15, "Berlin"], ["juli", 14, "NY"]]
df = pd.DataFrame(data, columns = ["name", "age", "city"])
df.set_index("name")
v3c.write(backend="kv", table="mytable", dfs=tsdf, expression="city="NY"", condition="age>14")
```

#### Read

Reads data from a backend.

Common read parameters:

- iterator: bool - Return iterator of DataFrames or (if False) just one DataFrame
- filter: string - Query filter (can't be used with query)
- columns: []str - List of columns to pass (can't be used with query)
- data_format: string - Data format (Not yet supported)
- marker: string - Query marker (Not yet supported)
- limit: int - Maximal number of rows to return (Not yet supported)
- row_layout: bool - Weather to use row layout (vs the default column layout) (Not yet supported)

##### Backend-Specific Parameters

###### TSDB

- start: string
- end: string
- step: string
- aggregators: string
- aggregationWindow: string
- query: string - Query in SQL format
- group_by: string - Query group by (can't be used with query)
- multi_index: bool - Get the results as a multi index data frame where the labels are used as indexes
 in addition to the timestamp, or if `False` (default behavior) only the timestamp will function as the index.

For detailed info on these parameters please visit [TSDB](https://github.com/v3io/v3io-tsdb#v3io-tsdb) docs.

Example:
```python
df = client.read(backend="tsdb", query="select avg(cpu) as cpu, avg(diskio), avg(network)from mytable", start="now-1d", end="now", step="2h")
```

###### NoSQL

- reset_index: bool - Reset the index. When set to `false` (default), the dataframe will have the key column of the v3io kv as the index column.
  When set to `true`, the index will be reset to a range index.
- max_in_message: int - Maximal number of rows per message
- sharding_keys: []string (Experimental)- list of specific sharding keys to query. For range scan formatted tables only.
- segments: []int64 (Not yet supported)
- total_segments: int64 (Not yet supported)
- sort_key_range_start: string (Not yet supported)
- sort_key_range_end: string (Not yet supported)

For detailed information on these parameters, refer to the platform's NoSQL documentation.

Example:
```python
df = client.read(backend="kv", table="mytable", filter="col1>666")
```

###### Stream

- seek: string - valid values:  `time | seq/sequence | latest | earliest`.
  <br/>
  if `seq` seek type is requested, need to provide the desired sequence id via `sequence` parameter.
  <br/>
  if `time` seek type is requested, need to provide the desired start time via `start` parameter.
- shard_id: string
- sequence: int64 (optional)

For detailed info on these parameters please visit [Stream](https://www.iguazio.com/docs/concepts/latest-release/streams) docs.

Example:
```python
df = client.read(backend="stream", table="mytable", seek="latest", shard_id="5")
```

#### Delete

Deletes a table of a specific backend.

Example:
```python
df = client.delete(backend="<backend>", table="mytable")
```

##### Backend-Specific Parameters

###### TSDB

- start: string - delete since start
- end: string - delete since start

Note: if both `start` and `end` are not specified **all** the TSDB table will be deleted.
For detailed information on these parameters, refer to the [V3IO TSDB](https://github.com/v3io/v3io-tsdb#v3io-tsdb) documentation.

Example:
```python
df = client.delete(backend="tsdb", table="mytable", start="now-1d", end="now-5h")
```

###### NoSQL

- filter: string - Filter for selective delete

Example:
```python
df = client.delete(backend="kv", table="mytable", filter="age>40")
```

#### Execute

Provides additional functions that are not covered in the basic CRUD functionality.

###### TSDB

Currently no `execute` commands are available for the TSDB backend.

###### NoSQL

- infer, inferschema - inferring and creating a schema file for a given kv table.
  <br/>
  Example: `client.execute(backend="kv", table="mytable", command="infer")`
- update - perform an update expression for a specific key.
  <br/>
  Example: `client.execute(backend="kv", table="mytable", command="update", args={"key": "somekey", "expression": "col2=30", "condition": "col3>15"})`

###### Stream

- put - putting a new object to a stream.
  <br/>
  Example: `client.execute(backend="stream", table="mystream", command="put", args={"data": "this a record", "clientinfo": "some_info", "partition": "partition_key"})`

## Contributing

To contribute to V3IO Frames, you need to be aware of the following:

- [Components](#components)
- [Development](#development)
  - [Adding and Changing Dependencies](#adding-and-changing-dependencies)
  - [Travis CI](#travis-ci)
- [Docker Image](#docker-image)
  - [Building the Image](#building-the-image)
  - [Running the Image](#running-the-image)

### Components

- Go server with support for both the gRPC and HTTP protocols
- Go client
- Python client

### Development

The core is written in [Go](https://golang.org/).
The development is done on the `development` branch and then released to the `master` branch.

- To execute the Go tests, run `make test`.
- To execute the Python tests, run `make test-python`.

#### Adding and Changing Dependencies

- If you add Go dependencies, run `make update-go-deps`.
- If you add Python dependencies, update **clients/py/Pipfile** and run `make
  update-py-deps`.

#### Travis CI

Integration tests are run on [Travis CI](https://travis-ci.org/).
See **.travis.yml** for details.

The following environment variables are defined in the [Travis settings](https://travis-ci.org/v3io/frames/settings):

- Docker Container Registry ([Quay.io](https://quay.io/))
    - `DOCKER_PASSWORD` &mdash; Password for pushing images to Quay.io.
    - `DOCKER_USERNAME` &mdash; Username for pushing images to Quay.io.
- Python Package Index ([PyPI](https://pypi.org/))
    - `V3IO_PYPI_PASSWORD` &mdash; Password for pushing a new release to PyPi.
    - `V3IO_PYPI_USER` &mdash; Username for pushing a new release to PyPi.
- Iguazio Data Science Platform
    - `V3IO_SESSION` &mdash; A JSON encoded map with session information for running tests.
      For example:

      ```
      '{"url":"45.39.128.5:8081","container":"mitzi","user":"daffy","password":"rabbit season"}'
      ```
      > **Note:** Make sure to embed the JSON object within single quotes (`'{...}'`).

### Docker Image

#### Building the Image

Use the following command to build the Docker image:

```sh
make build-docker
```

#### Running the Image

Use the following command to run the Docker image:

```sh
docker run \
	-v /path/to/config.yaml:/etc/framesd.yaml \
	quay.io/v3io/frames:unstable
```

## LICENSE

[Apache 2](LICENSE)

