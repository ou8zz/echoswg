# echoswg

#### 升级了 echo 4.12
#### 升级了 goalng  1.22

#### 实现了对swagger的参数返回统一并且返回值interface的tag标记输出

util.ResultData的实现，对controller的返回值进行统一处理
```
type (
  ResponseData struct {
    Errno int         `json:"errno"`          // 必需 错误码。正常返回0 异常返回560 错误提示561对应errorInfo
    Data  interface{} `json:"data,string"`    // 必需 返回数据内容。 如果有返回数据，可以是字符串或者数组JSON等等
    Page  interface{} `json:"page,omitempty"` // 非必需 分页信息
  }
)

// 固定参数要求0:数据结构，1:分页(如有)，error在最后
func ResultData(data ...interface{}) interface{} {
  result := &ResponseData{}
  for _, v := range data {
    switch v.(type) {
    case error:
      result.Errno = 501
      if bizError, ok := v.(*BizError); ok {
        result.Errno = bizError.code
        result.Data = v
      }
      AppLog.With("error", "ResultData").Errorf("error:%v", v)
      return result
    case Pagination:
      result.Page = v
    default:
      if v != nil {
        result.Data = v
      }
    }
  }
  return result
}
```