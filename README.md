# Logging Container

Logging Container is a system container that gathers logs and provides REST API for adding and querying logs from containers. 

Use Logger Container if you want to gather all the logs and find them easily. 

### Config:

The config for the container should look the following way:

```
{"access_tokens" : ["Some_Access_Token"], "cleanfrequency" : "1h40m", "ttl" : "24h"}
```

Config explanation:

*"access_tokens"* - required. array of strings, each is acceptable access token that is allowed to interact with logging service

*"cleanfrequency"* - optional. string, frequency of database cleanup. Parameter is a sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

*"ttl"* - optional. string, log entry time to live. Parameter is a sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".


### Post methods


**Add logs Endpoint (Post)**

```localhost:[mapped_port]/logs/add```


Use access-token parameter provided from container config in Headers.


Request body example:

```
{
   "publisher": "sasha",
   "timestamp": 149622569742,
   "level": "WARNING",
   "message": "sample message of level WARNING"
}
```

Response:

```Status 200 OK```

POST parameters description

*"publisher"*- (Text)

*"timestamp"*: the timestamp of container logs (Number)

*"level"*: string log level ("CRITICAL", "FATAL", "ERROR", "WARN", "WARNING", "INFO", "DEBUG")

*"message"*- (Text)


**Get logs Endpoint (Post)**

```localhost:[mapped_port]/logs/get```


Use access-token parameter provided from container config in Headers.


Request body example:

```
{
 "level" : "INFO",
 "publishers" : ["sasha", "stani"],
 "page" : 2,
 "pagesize":100,
 "timeframestart" : 149622500000,
 "timeframeend" : 149622599999,
 "orderby" : ["timestamp"],
 "asc" : false,
 "message":"sample mess"
}
```

Response example:

```
{

 "logs": [
  {
   "publisher": "sasha",
   "timestamp": 149622569742,
   "level": "WARNING",
   "message": "sample message of level WARNING"
  },
  {
   "publisher": "stani",
   "timestamp": 149622569728,
   "level": "ERROR",
   "message": "sample message of level ERROR"
  },
...
 ]
}
```

POST parameters description

All these parameters are optional. An empty request will return ALL the records.

*"timeframestart"* - time in mills

*"timeframeend"* - time in mills

*"publishers"* - array of publishers id

*"level"* - string log level ("CRITICAL", "FATAL", "ERROR", "WARN", "WARNING", "INFO", "DEBUG"), if ommited any level will be returned; all the higher levels for defined level are returned, e.g. for defined "ERROR" level "CRITICAL" and "FATAL" levels are also returned

*"message"* - string to search for (sql will be LIKE %message%)

*"page"* - number of page to return, default to 1

*"orderby"* - array of feild to order by (any of "publisher" "level" "message" "timestamp"), default to timestamp column only

*"asc"* - sort orfer - true/false- default to false

*"pagesize"* - page size; max number of logs in one page to return, default to 20
