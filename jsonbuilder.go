package github.com/steviesama/jsonbuilder

import (
  "fmt"
  "strings"
  "reflect"
)

type JsonBuilder struct {
	Data string
	Level []string
	CurrentDepth int
}

func (j *JsonBuilder) getIndent() string {
	indentation := ""
	for i := 0; i < j.CurrentDepth; i++ { indentation += "\t"}
	return indentation
}

func (j *JsonBuilder) StartObject() {
	j.Data += j.getIndent() + "{\n"
	j.CurrentDepth++
	j.Level = append(j.Level, "}")
}

func (j *JsonBuilder) StartArray() {
	j.Data += j.getIndent() + "[\n"
	j.CurrentDepth++
	j.Level = append(j.Level, "]")
}

func (j *JsonBuilder) EndLevel() {
	if j.CurrentDepth == 0 {
		fmt.Println("JsonBuilder: End() called at CurrentDepth == 0")
		return
	}
	j.CurrentDepth--
	j.Data += fmt.Sprintf("%s%s\n", j.getIndent(), j.Level[len(j.Level) - 1])
	j.Level = j.Level[:len(j.Level) - 1]
}

func (j *JsonBuilder) EndAllLevels() {
  for i := len(j.Data) - 1; i < 0; i-- {
    j.CurrentDepth--
  	j.Data += fmt.Sprintf("%s%s\n", j.getIndent(), j.Level[i])
  }
	j.Level = make([]string, 0)
}

func (j *JsonBuilder) GetLine(jsonVarName string, value interface{}, isLastLine bool) string {
	indentation := ""
	var commaString string
	var data string
	if !isLastLine { commaString = ","
	} else { commaString = "" }
	for i := 0; i < j.CurrentDepth; i++ { indentation += "\t"}
	data += fmt.Sprintf("%s\"%s\": ", indentation, jsonVarName)
	switch value.(type) {
		case string:
			return data + fmt.Sprintf("\"%s\"%s\n", value, commaString)
		case float32:
      return data + fmt.Sprintf("\"%f\"%s\n", value, commaString)
    case float64:
			return data + fmt.Sprintf("\"%f\"%s\n", value, commaString)
		default:
			return data + fmt.Sprintf("%d%s\n", value, commaString)
	}
}

func (j *JsonBuilder) Add(jsonVarName string, value interface{}) {
	indentation := ""
	for i := 0; i < j.CurrentDepth; i++ { indentation += "\t"}
	j.Data += fmt.Sprintf("%s\"%s\": ", indentation, jsonVarName)
	switch value.(type) {
		case string:
			j.Data += fmt.Sprintf("\"%s\",\n", value)
		default:
			j.Data += fmt.Sprintf("%d,\n", value)
	}
}

func (j *JsonBuilder) AddLast(jsonVarName string, value interface{}) {
	indentation := ""
	for i := 0; i < j.CurrentDepth; i++ { indentation += "\t"}
	j.Data += fmt.Sprintf("%s\"%s\": ", indentation, jsonVarName)
	switch value.(type) {
		case string:
			j.Data += fmt.Sprintf("\"%s\"\n", value)
		default:
			j.Data += fmt.Sprintf("%d\n", value)
	}
	j.EndLevel()
}

func NewJsonBuilder() JsonBuilder {
	return JsonBuilder{Data: "", CurrentDepth: 0, Level: make([]string, 0)}
}

func isStruct(object interface{}) bool {
	v := reflect.Indirect(reflect.ValueOf(object))
	switch v.Kind() {
		case reflect.Struct:
			// fmt.Printf("\n-----\nstruct.Interface(): %v\n-----\n\n", v.Interface())
			return true
	}
	return false
}

func numFields(object interface{}) int {
	if isStruct(object) {
		// fmt.Printf("\n-----\nobject: %v\n-----\n\n", object)
		v := reflect.Indirect(reflect.ValueOf(object))
		return v.Type().NumField() // reflect.TypeOf(object).NumField()
	}
	return -1
}

func GetJsonFlattenedLinesFunc() (getJsonFlattenedLines (func(interface{}) string)) {
	recursionDepth := 0
	json := NewJsonBuilder()
	json.StartObject()
	data := ""
  var isLastStruct bool
	getJsonFlattenedLines = func(object interface{}) string {
		recursionDepth++
		interpretLastLine := recursionDepth == 1
		v := reflect.Indirect(reflect.ValueOf(object))
		fieldCount := numFields(object)
			for i := 0; i < fieldCount; i++ {
        if(recursionDepth == 1) {
          isLastStruct = i == fieldCount - 1
        }
				isLastLine := i == fieldCount - 1
				if !interpretLastLine && isLastLine && !isLastStruct {
					isLastLine = false
					// fmt.Printf("!interpretLastLine && isLastLine\n")
				}
				switch v.Field(i).Kind() {
					case reflect.Struct:
						// fmt.Printf("fieldName: %s\n", fmt.Sprintf("%v", v.Type().Field(i).Name))
						data += getJsonFlattenedLines(v.Field(i).Interface())
						// fmt.Println("reflect.Struct")
					default:
						// fmt.Printf("fieldName: %s\n", fmt.Sprintf("%v", v.Type().Field(i).Name))
						tag := GetJsonTag(object, fmt.Sprintf("%v", v.Type().Field(i).Name))
						data += json.GetLine(tag, v.Field(i).Interface(), isLastLine)
						// fmt.Println("default")
				}
			}
			recursionDepth--

			if recursionDepth == 0 && isLastStruct{
				json.Data += data
				json.EndLevel()
				return json.Data
			} else { return data }
	}
	return
}

func GetJson(object interface{}) string {
    getJson := GetJsonFlattenedLinesFunc()
    return getJson(object)
}

func GetJsonTag(i interface{}, fieldName string) string {
	// fmt.Printf("\n\n-----\nGetJsonTag.i == '%v'\n-----\n\n", i)
	v := reflect.Indirect(reflect.ValueOf(i))
	// fmt.Printf("\n\n-----\nv.Type() == '%v'\n-----\n\n", v.Type())
	field, ok := v.Type()/*.Elem() don't use Elem() if it's not an interface{} or ptr*/.FieldByName(fieldName)
	if !ok {
		fmt.Printf("GetJsonTag(): reflect.TypeOf(i).Elem().FieldByName(%s) failed\n", fieldName)
		return ""
	}
	tag := strings.Split(strings.Split(string(field.Tag), ":")[1], "\"")[1]
	return tag
}
