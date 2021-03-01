# API Common 

All collection endpoints have:
 * Filtering based on equality: `?property1=value1&property1=value2&property2=value3`
 * Sorting: `?sortBy=property:asc`/`?sortBy=property:desc` 
 * Pagination: 
   * In order to use pagination, a combination of parameters have to used:
        * get first 100 items: `?sortBy=property:asc&count=100`
        * get next 100 items: `?sortBy=property:asc&count=100&lastValue=value` where `lastValue` is the value of the sort property of the last returned item


## Endpoints:
  * [Executions](executions.md)
  * [Projects](projects.md)
  * [Scenarios](scenarios.md)
  * [Test Plans](testplans.md)