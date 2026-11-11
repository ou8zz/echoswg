package echoswg

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"reflect"
)

type RespObj struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data,string"`
}

func genRespObjType(dataType reflect.Type) reflect.Type {
	respObjType := reflect.TypeOf(RespObj{})
	if dataType == nil {
		return respObjType
	}
	fields := make([]reflect.StructField, respObjType.NumField())
	for i := 0; i < respObjType.NumField(); i++ {
		field := respObjType.Field(i)
		if field.Name == "Data" {
			field.Type = dataType
		}
		fields[i] = field
	}
	return reflect.StructOf(fields)
}

func respJson(ctx *echo.Context, outs []reflect.Value) error {
	var rest RespObj
	if len(outs) > 0 {
		rest.Data = outs[0].Interface()
	}
	return (*ctx).JSON(http.StatusOK, rest)
}
