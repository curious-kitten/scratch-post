# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [execution.proto](#execution.proto)
    - [Execution](#metadata.scratchpost.curiouskitten.Execution)
    - [StepExecution](#metadata.scratchpost.curiouskitten.StepExecution)
  
    - [Status](#metadata.scratchpost.curiouskitten.Status)
  
- [Scalar Value Types](#scalar-value-types)



<a name="execution.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## execution.proto



<a name="metadata.scratchpost.curiouskitten.Execution"></a>

### Execution
Represents an execution of a scenario. It associates with a Scenario through the `scenarioId`. 
It needs an association with a project and a test plan. This is done through the `projectId` and `testPlanId`


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| identity | [Identity](#metadata.scratchpost.curiouskitten.Identity) |  | Identification for the execution |
| projectId | [string](#string) |  | ID of the associated project |
| scenarioId | [string](#string) |  |  |
| testPlanId | [string](#string) |  |  |
| status | [Status](#metadata.scratchpost.curiouskitten.Status) |  |  |
| name | [string](#string) |  |  |
| description | [string](#string) |  |  |
| prerequisites | [string](#string) |  |  |
| steps | [StepExecution](#metadata.scratchpost.curiouskitten.StepExecution) | repeated |  |
| issues | [LinkedIssue](#metadata.scratchpost.curiouskitten.LinkedIssue) | repeated |  |
| labels | [string](#string) | repeated |  |






<a name="metadata.scratchpost.curiouskitten.StepExecution"></a>

### StepExecution
Represents a step that has to be completed in order to complete the test


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| definition | [scenario.scratchpost.curiouskitten.Step](#scenario.scratchpost.curiouskitten.Step) |  | Definition of the step to be executed |
| status | [Status](#metadata.scratchpost.curiouskitten.Status) |  | Status of the execution. Defaults to Pending |
| ActualResult | [string](#string) |  | Details about the exectuion results |
| issues | [LinkedIssue](#metadata.scratchpost.curiouskitten.LinkedIssue) | repeated | Issues associated with the step execution |





 


<a name="metadata.scratchpost.curiouskitten.Status"></a>

### Status
Status of an execution

| Name | Number | Description |
| ---- | ------ | ----------- |
| Pending | 0 | an execution that has not been completed |
| Fail | 1 | an execution result did not match the expected |
| Pass | 2 | an execution result matches the expected |


 

 

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

