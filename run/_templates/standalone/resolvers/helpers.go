package resolvers

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Returns a map of a struct's field names in all caps with their field positions, given an object of that struct
func StructFieldCapsNameMap[T any](obj *T) (map[string]int, error) {
	val := reflect.ValueOf(obj)
	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return nil, errors.New("must provide a struct")
	}

	m := make(map[string]int)

	for i := 0; i < elem.NumField(); i++ {
		name := elem.Type().Field(i).Name
		m[strings.ToUpper(name)] = i
	}

	return m, nil
}

func ValOrPointerVal(val reflect.Value) reflect.Value {
	if val.Kind() == reflect.Ptr {
		return val.Elem()
	}
	return val
}

func GoStructToPb[G any, P any](gostruct G) P {
	var pb P
	pbv := reflect.ValueOf(&pb).Elem()

	gostructv := reflect.ValueOf(&gostruct).Elem()

	gostructmap, err := StructFieldCapsNameMap(&gostruct)
	if err != nil {
		panic(err)
	}

	pbmap, err := StructFieldCapsNameMap(&pb)
	if err != nil {
		panic(err)
	}

	for key, val := range pbmap {
		if key == "STATE" || key == "SIZECACHE" || key == "UNKNOWNFIELDS" || key == "ID" {
			continue
		}
		if _, ok := gostructmap[key]; ok {
			gostructindex := gostructmap[key]
			switch pbv.Field(val).Type().String() {
			case "int32":
				v := ValOrPointerVal(gostructv.Field(gostructindex))
				pbv.Field(val).SetInt(v.Int())
			case "string":
				v := ValOrPointerVal(gostructv.Field(gostructindex))
				pbv.Field(val).SetString(v.String())
			case "bool":
				v := ValOrPointerVal(gostructv.Field(gostructindex))
				pbv.Field(val).SetBool(v.Bool())
			case "uint32":
				v := ValOrPointerVal(gostructv.Field(gostructindex))
				pbv.Field(val).Set(v.Convert(pbv.Field(val).Type()))
			case "*timestamppb.Timestamp":
				v := ValOrPointerVal(gostructv.Field(gostructindex))
				t, err := time.Parse(time.RFC3339, v.String())
				if err != nil {
					panic(err)
				}
				pbTime := timestamppb.New(t)
				pbv.Field(val).Set(reflect.ValueOf(pbTime))
			case "*wrapperspb.BoolValue":
				if gostructv.Field(gostructindex).IsNil() {
					continue
				}
				v := reflect.Indirect(gostructv.Field(gostructindex)).Bool()
				pbv.Field(val).Set(reflect.ValueOf(wrapperspb.Bool(v)))
			}
		} else {
			panic("cannot find field " + key + " in pb struct")
		}
	}
	return pb
}

func PbToGoStruct[P any, G any](pb P, skipId bool) G {
	var gostruct G
	gostructv := reflect.ValueOf(&gostruct).Elem()
	pbv := reflect.ValueOf(&pb).Elem()
	pbmap, err := StructFieldCapsNameMap(&pb)
	if err != nil {
		panic(err)
	}
	gostructmap, err := StructFieldCapsNameMap(&gostruct)
	if err != nil {
		panic(err)
	}
	// try and find all the values of the graphql struct
	for key, val := range gostructmap {
		if key == "STATE" || key == "SIZECACHE" || key == "UNKNOWNFIELDS" {
			continue
		}
		if skipId && key == "ID" {
			continue
		}
		if _, ok := pbmap[key]; ok {
			pbindex := pbmap[key]
			switch gostructv.Field(val).Type().String() {
			case "int32":
				v := ValOrPointerVal(pbv.Field(pbindex))
				gostructv.Field(val).SetInt(v.Int())
			case "*int32":
				v := int32(pbv.Field(pbindex).Uint())
				gostructv.Field(val).Set(reflect.ValueOf(&v))
			case "graphql.ID":
				v := ValOrPointerVal(pbv.Field(pbindex))
				gostructv.Field(val).SetString(v.String())
			case "string":
				switch ValOrPointerVal(pbv.Field(pbindex)).String() {
				case "<timestamppb.Timestamp Value>":
					s := ValOrPointerVal(pbv.Field(pbindex))
					getFieldSecond := s.FieldByName("Seconds")
					t := time.Unix(getFieldSecond.Int(), 0).UTC()
					formattedTime := t.Format(time.RFC3339)
					gostructv.Field(val).Set(reflect.ValueOf(&formattedTime))
				default:
					v := ValOrPointerVal(pbv.Field(pbindex))
					gostructv.Field(val).SetString(v.String())
				}
			case "*string":
				switch ValOrPointerVal(pbv.Field(pbindex)).String() {
				case "<timestamppb.Timestamp Value>":
					s := ValOrPointerVal(pbv.Field(pbindex))
					getFieldSecond := s.FieldByName("Seconds")
					t := time.Unix(getFieldSecond.Int(), 0).UTC()
					formattedTime := t.Format(time.RFC3339)
					gostructv.Field(val).Set(reflect.ValueOf(&formattedTime))
				default:
					v := pbv.Field(pbindex).String()
					gostructv.Field(val).Set(reflect.ValueOf(&v))
				}

			case "bool":
				v := ValOrPointerVal(pbv.Field(pbindex))
				gostructv.Field(val).SetBool(v.Bool())
			case "*bool":
				v := pbv.Field(pbindex).Bool()
				gostructv.Field(val).Set(reflect.ValueOf(&v))
			}
		} else {
			panic("cannot find field " + key + " in pb struct")
		}
	}
	return gostruct
}
