package echoswg

import (
  "reflect"
  "strings"
  "fmt"
  "time"
)

var GlobalTypeDefBuilder = NewTypeDefBuilder()
type TypeDefBuilder struct {
  cachedTypes []reflect.Type
  position int
  StructDefinitions map[string]map[string]interface{}
  typeNames map[reflect.Type]string
  //anonymousTypes map[reflect.Type]string
}

func NewTypeDefBuilder() *TypeDefBuilder {
  return &TypeDefBuilder {
    cachedTypes: make([]reflect.Type, 0),
    position: 0,
    StructDefinitions: make(map[string]map[string]interface{}),
    typeNames: make(map[reflect.Type]string),
  }
}

func (b *TypeDefBuilder) Build(typ reflect.Type) *SwaggerType {
  swaggerType := b.ToSwaggerType(typ)

  for b.position < len(b.cachedTypes) {
    pendingType := b.cachedTypes[b.position]
    typeName := b.uniqueStructName(pendingType)
    if _, ok := b.StructDefinitions[typeName]; !ok {
      b.StructDefinitions[typeName] = propertiesOfEntity(pendingType)
    }
    b.position += 1
  }
  return swaggerType
}

func propertiesOfEntity(bodyType reflect.Type) map[string]interface{} {
  fmt.Printf("propertiesOfEntity: %s\n", bodyType)
  properties := make(map[string]interface{})
  requiredFields := []string{}
  for i := 0; i < bodyType.NumField(); i++ {
    field := bodyType.Field(i)
    propertyName := field.Name
    propertyJsonName := strings.Split(field.Tag.Get("json"), ",")[0]
    if len(propertyJsonName)> 0 {
      propertyName = propertyJsonName
    }
    swaggerType := GlobalTypeDefBuilder.ToSwaggerType(field.Type)

    if !swaggerType.Optional {
      requiredFields = append(requiredFields, propertyName)
    }

    propertyJson := swaggerType.ToSwaggerJSON()

    description := field.Tag.Get("desc")
    description = strings.TrimSpace(description)
    if len(description) > 0 {
      propertyJson["description"] = description
    }

    properties[propertyName] = propertyJson
  }
  return map[string]interface{}{
    "type":       "object",
    "required":   requiredFields,
    "properties": properties,
  }
}


func (b *TypeDefBuilder) uniqueStructName(typ reflect.Type) string {
  if existed, ok := b.typeNames[typ]; ok {
    return  existed
  }

  typeName := typ.Name()
  if len(typeName) == 0 {
    typeName = "anonymous"
  }

  //typeName = fmt.Sprintf("anonymous%02d", len(b.anonymousTypes))
  var getNameSuccess = false
  var existedCount = 0
  for !getNameSuccess {
    var isExisted = false
    for _, name := range b.typeNames {
      if name == typeName {
        isExisted = true
        existedCount += 1
        break
      }
    }
    if isExisted {
      typeName = fmt.Sprintf("%s%02d", typeName, existedCount)
    } else {
      getNameSuccess = true
    }
  }

  b.typeNames[typ] = typeName
  return typeName
}

type SwaggerType struct {
  Optional bool
  Type string
  Format string
  Items *SwaggerType
}

func (t *SwaggerType) String() string {
  if t == nil {
    return ""
  }
  switch t.Type {
  case "array": return fmt.Sprintf("type: array, items: [%s]", t.Items.String())
  case "object": return fmt.Sprintf("$ref: %s", t.Format)
  default:
    return fmt.Sprintf("optional: %t, type: %s, format: %s", t.Optional, t.Type, t.Format)
  }
}

func (t *SwaggerType) ToSwaggerJSON() map[string]interface{} {
  switch t.Type {
  case "array":
    return map[string]interface{} {
      "type": "array",
      "items": t.Items.ToSwaggerJSON(),
    }
  case "object": return map[string]interface{} {
    "$ref": t.Format,
  }
  default:
   return map[string]interface{} {
      "type": t.Type,
      "format": t.Format,
     }
  }
}
func (b *TypeDefBuilder) ToSwaggerType(typ reflect.Type) *SwaggerType {
  v := &SwaggerType{}
  b._toSwaggerType(typ, v)
  return v
}

var TimeType = reflect.TypeOf((*time.Time)(nil)).Elem()
func (b *TypeDefBuilder) _toSwaggerType(typ reflect.Type, dest *SwaggerType) {
  if typ == TimeType {
    dest.Type = "string"
    dest.Format = "string"
    return
  }
  switch typ.Kind() {
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
    reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
    dest.Type = "integer"
    dest.Format = "int32"
    return
  case reflect.Int64, reflect.Uint64:
    dest.Type = "integer"
    dest.Format = "int64"
    return
  case reflect.String:
    dest.Type = "string"
    dest.Format = "string"
    return
  case reflect.Float32:
    dest.Type = "number"
    dest.Format = "float"
    return
  case reflect.Float64:
    dest.Type = "number"
    dest.Format = "double"
    return
  case reflect.Bool:
    dest.Type = "boolean"
    dest.Format = "boolean"
    return
  case reflect.Array, reflect.Slice:
    dest.Type = "array"
    itemType := &SwaggerType{}
    b._toSwaggerType(typ.Elem(), itemType)
    dest.Items = itemType
    return
  case reflect.Ptr:
    dest.Optional = true
    b._toSwaggerType(typ.Elem(), dest)
    return
  case reflect.Struct:
    dest.Type = "object"
    dest.Format = "#/definitions/" + b.uniqueStructName(typ)
    b.cachedTypes = append(b.cachedTypes, typ)
    //fmt.Printf("add type to cache: %s", typ)
    return
  default:
    return
  }
}
