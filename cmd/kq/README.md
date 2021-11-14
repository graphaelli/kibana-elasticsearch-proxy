# Kibana Query Tool

By example:

Create an index and add some documents:
```shell
DELETE foo
PUT foo
{
  "mappings": {
    "properties": {
      "@timestamp": {"type": "date"}
    }
  }
}
POST foo/_bulk
{"index": {}}
{"@timestamp":"2021-11-11T18:00:00.000Z", "a": 1}
{"index": {}}
{"@timestamp":"2021-11-11T18:00:00.000Z", "a": 2}
{"index": {}}
{"@timestamp":"2021-11-11T18:00:00.000Z", "a": 3}
{"index": {}}
{"@timestamp":"2021-11-11T18:00:00.000Z", "a": 4}
{"index": {}}
{"@timestamp":"2021-11-11T18:00:00.000Z", "a": 5}
{"index": {}}
{"@timestamp":"2021-11-11T18:00:00.000Z", "a": 6}
{"index": {}}
{"@timestamp":"2021-11-11T18:00:00.000Z", "a": 7}
{"index": {}}
{"@timestamp":"2021-11-11T18:00:00.000Z", "a": 8}
{"index": {}}
{"@timestamp":"2021-11-11T18:00:00.000Z", "a": 9}
```

Pull them out a few at a time including auth and a filter path:
```shell
go run ./cmd/kq -u http://admin:changeme@localhost:5601/ -index foo -p 6 -f hits.hits._source.a,hits.hits.sort,hits.hits.pit,pit_id -q '
{
  "size": 2,
  "query": {
    "bool": {
      "filter": [
        {
          "range": {
            "@timestamp": {
              "gt": "now-168h/h"
            }
          }
        }
      ]
    }
  },
  "sort": [
    {
      "@timestamp": {
        "order": "asc",
        "format": "strict_date_optional_time_nanos"
      }
    },
    {
      "_shard_doc": "asc"
    }
  ],
  "_source": [
    "@timestamp",
    "a"
  ]
}'
```

Resulting in:
```shell
{
  "pit_id" : "o4K1AwEDZm9vFnM3Smk4dloyUlJXeFlBeU1IZ2ZhRHcAFlNRYy02bjV6U3h1MGNUek9GQVVsN2cAAAAAAAAK5BcWUmhiNEhpOWNTa3lqSmU0bEt6Zlh2ZwABFnM3Smk4dloyUlJXeFlBeU1IZ2ZhRHcAAA==",
  "hits" : {
    "hits" : [
      {
        "_source" : {
          "a" : 1
        },
        "sort" : [
          "2021-11-11T18:00:00.000Z",
          0
        ]
      },
      {
        "_source" : {
          "a" : 2
        },
        "sort" : [
          "2021-11-11T18:00:00.000Z",
          1
        ]
      }
    ]
  }
}
{
  "pit_id" : "o4K1AwEDZm9vFnM3Smk4dloyUlJXeFlBeU1IZ2ZhRHcAFlNRYy02bjV6U3h1MGNUek9GQVVsN2cAAAAAAAAK5BcWUmhiNEhpOWNTa3lqSmU0bEt6Zlh2ZwABFnM3Smk4dloyUlJXeFlBeU1IZ2ZhRHcAAA==",
  "hits" : {
    "hits" : [
      {
        "_source" : {
          "a" : 3
        },
        "sort" : [
          "2021-11-11T18:00:00.000Z",
          2
        ]
      },
      {
        "_source" : {
          "a" : 4
        },
        "sort" : [
          "2021-11-11T18:00:00.000Z",
          3
        ]
      }
    ]
  }
}
{
  "pit_id" : "o4K1AwEDZm9vFnM3Smk4dloyUlJXeFlBeU1IZ2ZhRHcAFlNRYy02bjV6U3h1MGNUek9GQVVsN2cAAAAAAAAK5BcWUmhiNEhpOWNTa3lqSmU0bEt6Zlh2ZwABFnM3Smk4dloyUlJXeFlBeU1IZ2ZhRHcAAA==",
  "hits" : {
    "hits" : [
      {
        "_source" : {
          "a" : 5
        },
        "sort" : [
          "2021-11-11T18:00:00.000Z",
          4
        ]
      },
      {
        "_source" : {
          "a" : 6
        },
        "sort" : [
          "2021-11-11T18:00:00.000Z",
          5
        ]
      }
    ]
  }
}
{
  "pit_id" : "o4K1AwEDZm9vFnM3Smk4dloyUlJXeFlBeU1IZ2ZhRHcAFlNRYy02bjV6U3h1MGNUek9GQVVsN2cAAAAAAAAK5BcWUmhiNEhpOWNTa3lqSmU0bEt6Zlh2ZwABFnM3Smk4dloyUlJXeFlBeU1IZ2ZhRHcAAA==",
  "hits" : {
    "hits" : [
      {
        "_source" : {
          "a" : 7
        },
        "sort" : [
          "2021-11-11T18:00:00.000Z",
          6
        ]
      },
      {
        "_source" : {
          "a" : 8
        },
        "sort" : [
          "2021-11-11T18:00:00.000Z",
          7
        ]
      }
    ]
  }
}
{
  "pit_id" : "o4K1AwEDZm9vFnM3Smk4dloyUlJXeFlBeU1IZ2ZhRHcAFlNRYy02bjV6U3h1MGNUek9GQVVsN2cAAAAAAAAK5BcWUmhiNEhpOWNTa3lqSmU0bEt6Zlh2ZwABFnM3Smk4dloyUlJXeFlBeU1IZ2ZhRHcAAA==",
  "hits" : {
    "hits" : [
      {
        "_source" : {
          "a" : 9
        },
        "sort" : [
          "2021-11-11T18:00:00.000Z",
          8
        ]
      }
    ]
  }
}
{
  "pit_id" : "o4K1AwEDZm9vFnM3Smk4dloyUlJXeFlBeU1IZ2ZhRHcAFlNRYy02bjV6U3h1MGNUek9GQVVsN2cAAAAAAAAK5BcWUmhiNEhpOWNTa3lqSmU0bEt6Zlh2ZwABFnM3Smk4dloyUlJXeFlBeU1IZ2ZhRHcAAA=="
}
```