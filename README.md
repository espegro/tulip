# tulip - Simple stable bloomfilter web service

A simple go web service implementing SBF

[Stable Learned Bloom Filters for Data Streams](http://www.vldb.org/pvldb/vol13/p2355-liu.pdf)

More about setting filter values: https://hur.st/bloomfilter/

## Usage
```
tulip -port 8080 -address 127.0.0.1 -state filterstate.json
```


## Methods

### POST

#### New
Create a new filter named __name__

* _filtersize_ number of cells in filter - m 
* _hashfunctions_ number of hash functions - k
* _decay_ number of random deletes before insert
* _max_ max cell value
  

```
/bloom/new/<name>/<filtersize>/<hashfunctions>/<decay>/<max>
```


#### Add
Add a single value to filter

```
/bloom/add/<name>/<value>
```

#### AddIfNotSet
Add single value to filter if value not exists in filter

```
/bloom/addifnotset/<name>/<value>
```

#### Poster
Add multiple values to filter

```
/bloom/poster/<name>
```

Eg. **filename.txt**
```
value1
value2
value3
```

```
curl -X POST http://127.0.0.1:8080/bloom/poster/filtername --data-binary @filename.txt
```
#### Reset
Set all values in filter to zero
```
/bloom/reset/<name>
```

#### Destroy
Remove filter
```
/bloom/destroy/<name>
```

### GET

#### Test
Check if value is present in filter
```
/bloom/test/<value>
```
#### List
List filters
```
/bloom/list
```

#### Info
Get info about filter
```
/bloom/info/<name>
```

#### Debug
Get debug info about filter
```
/bloom/debug/<name>
```

#### Save
Force saving of filterstate to disk
```
/bloom/save
```

#### Load
Force reload of filterstate from disk
```
/bloom/load
```


