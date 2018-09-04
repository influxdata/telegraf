// +build !as_performance

// Copyright 2013-2017 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"time"

	. "github.com/aerospike/aerospike-client-go/types"
	Buffer "github.com/aerospike/aerospike-client-go/utils/buffer"
)

// if this file is included in the build, it will include this method
func init() {
	objectParser = parseObject
}

func parseObject(
	cmd *readCommand,
	opCount int,
	fieldCount int,
	generation uint32,
	expiration uint32,
) error {
	receiveOffset := 0

	// There can be fields in the response (setname etc).
	// But for now, ignore them. Expose them to the API if needed in the future.
	// Logger.Debug("field count: %d, databuffer: %v", fieldCount, cmd.dataBuffer)
	if fieldCount > 0 {
		// Just skip over all the fields
		for i := 0; i < fieldCount; i++ {
			// Logger.Debug("%d", receiveOffset)
			fieldSize := int(Buffer.BytesToUint32(cmd.dataBuffer, receiveOffset))
			receiveOffset += (4 + fieldSize)
		}
	}

	if opCount > 0 {
		rv := *cmd.object

		if rv.Kind() != reflect.Ptr {
			return errors.New("Invalid type for result object. It should be of type Struct Pointer.")
		}
		rv = rv.Elem()

		if !rv.CanAddr() {
			return errors.New("Invalid type for object. It should be addressable (a pointer)")
		}

		if rv.Kind() != reflect.Struct {
			return errors.New("Invalid type for object. It should be a pointer to a struct.")
		}

		// find the name based on tag mapping
		iobj := indirect(rv)
		mappings := objectMappings.getMapping(iobj.Type())

		if err := setObjectMetaFields(iobj, TTL(expiration), generation); err != nil {
			return err
		}

		for i := 0; i < opCount; i++ {
			opSize := int(Buffer.BytesToUint32(cmd.dataBuffer, receiveOffset))
			particleType := int(cmd.dataBuffer[receiveOffset+5])
			nameSize := int(cmd.dataBuffer[receiveOffset+7])
			name := string(cmd.dataBuffer[receiveOffset+8 : receiveOffset+8+nameSize])
			receiveOffset += 4 + 4 + nameSize

			particleBytesSize := int(opSize - (4 + nameSize))
			value, _ := bytesToParticle(particleType, cmd.dataBuffer, receiveOffset, particleBytesSize)
			if err := setObjectField(mappings, iobj, name, value); err != nil {
				return err
			}

			receiveOffset += particleBytesSize
		}
	}

	return nil
}

func setObjectMetaFields(obj reflect.Value, ttl, gen uint32) error {
	// find the name based on tag mapping
	iobj := indirect(obj)

	ttlMap, genMap := objectMappings.getMetaMappings(iobj.Type())

	if ttlMap != nil {
		for i := range ttlMap {
			f := iobj.FieldByName(ttlMap[i])
			if err := setValue(f, ttl); err != nil {
				return err
			}
		}
	}

	if genMap != nil {
		for i := range genMap {
			f := iobj.FieldByName(genMap[i])
			if err := setValue(f, gen); err != nil {
				return err
			}
		}
	}

	return nil
}

func setObjectField(mappings map[string]string, obj reflect.Value, fieldName string, value interface{}) error {
	if value == nil {
		return nil
	}

	if name, exists := mappings[fieldName]; exists {
		fieldName = name
	}
	f := obj.FieldByName(fieldName)
	return setValue(f, value)
}

func setValue(f reflect.Value, value interface{}) error {
	// find the name based on tag mapping
	if f.CanSet() {
		if value == nil {
			if f.IsValid() && !f.IsNil() {
				f.Set(reflect.ValueOf(value))
			}
			return nil
		}

		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f.SetInt(int64(value.(int)))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			switch v := value.(type) {
			case uint8:
				f.SetUint(uint64(v))
			case uint16:
				f.SetUint(uint64(v))
			case uint32:
				f.SetUint(uint64(v))
			case uint64:
				f.SetUint(uint64(v))
			case uint:
				f.SetUint(uint64(v))
			default:
				f.SetUint(uint64(value.(int)))
			}
		case reflect.Float64, reflect.Float32:
			// if value has returned as a float
			if fv, ok := value.(float64); ok {
				f.SetFloat(fv)
			} else {
				// otherwise it is an old float64<->int64 marshalling type cast which needs to be set as int
				f.SetFloat(float64(math.Float64frombits(uint64(value.(int)))))
			}
		case reflect.String:
			rv := reflect.ValueOf(value.(string))
			if rv.Type() != f.Type() {
				rv = rv.Convert(f.Type())
			}
			f.Set(rv)
		case reflect.Bool:
			f.SetBool(value.(int) == 1)
		case reflect.Interface:
			if value != nil {
				f.Set(reflect.ValueOf(value))
			}
		case reflect.Ptr:
			switch f.Type().Elem().Kind() {
			case reflect.Int:
				tempV := int(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Uint:
				tempV := uint(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.String:
				tempV := string(value.(string))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Int8:
				tempV := int8(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Uint8:
				tempV := uint8(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Int16:
				tempV := int16(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Uint16:
				tempV := uint16(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Int32:
				tempV := int32(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Uint32:
				tempV := uint32(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Int64:
				tempV := int64(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Uint64:
				tempV := uint64(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Float64:
				// it is possible that the value is an integer set in the field
				// via the old float<->int64 type cast
				var tempV float64
				if fv, ok := value.(float64); ok {
					tempV = fv
				} else {
					tempV = math.Float64frombits(uint64(value.(int)))
				}

				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Bool:
				tempV := bool(value.(int) == 1)
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Float32:
				// it is possible that the value is an integer set in the field
				// via the old float<->int64 type cast
				var tempV64 float64
				if fv, ok := value.(float64); ok {
					tempV64 = fv
				} else {
					tempV64 = math.Float64frombits(uint64(value.(int)))
				}

				tempV := float32(tempV64)
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Interface:
				f.Set(reflect.ValueOf(&value))
			case reflect.Struct:
				// support time.Time
				if f.Type().Elem().PkgPath() == "time" && f.Type().Elem().Name() == "Time" {
					tm := time.Unix(0, int64(value.(int)))
					f.Set(reflect.ValueOf(&tm))
					break
				} else {
					valMap := value.(map[interface{}]interface{})
					// iteraste over struct fields and recursively fill them up
					if valMap != nil {
						newObjPtr := f
						if f.IsNil() {
							newObjPtr = reflect.New(f.Type().Elem())
						}
						theStruct := newObjPtr.Elem().Type()
						numFields := newObjPtr.Elem().NumField()
						for i := 0; i < numFields; i++ {
							// skip unexported fields
							fld := theStruct.Field(i)
							if fld.PkgPath != "" {
								continue
							}

							alias := fld.Name
							tag := strings.Trim(fld.Tag.Get(aerospikeTag), " ")
							if tag != "" {
								alias = tag
							}

							if valMap[alias] != nil {
								if err := setValue(reflect.Indirect(newObjPtr).FieldByName(fld.Name), valMap[alias]); err != nil {
									return err
								}
							}
						}

						// set the field
						f.Set(newObjPtr)
					}
				}
			} // switch ptr
		case reflect.Slice, reflect.Array:
			// BLOBs come back as []byte
			theArray := reflect.ValueOf(value)

			if f.Kind() == reflect.Slice {
				if f.IsNil() {
					f.Set(reflect.MakeSlice(reflect.SliceOf(f.Type().Elem()), theArray.Len(), theArray.Len()))
				} else if f.Len() < theArray.Len() {
					count := theArray.Len() - f.Len()
					f = reflect.AppendSlice(f, reflect.MakeSlice(reflect.SliceOf(f.Type().Elem()), count, count))
				}
			}

			for i := 0; i < theArray.Len(); i++ {
				if err := setValue(f.Index(i), theArray.Index(i).Interface()); err != nil {
					return err
				}
			}
		case reflect.Map:
			emptyStruct := reflect.ValueOf(struct{}{})
			theMap := value.(map[interface{}]interface{})
			if theMap != nil {
				newMap := reflect.MakeMap(f.Type())
				var newKey, newVal reflect.Value
				for key, elem := range theMap {
					if key != nil {
						newKey = reflect.ValueOf(key)
					} else {
						newKey = reflect.Zero(f.Type().Key())
					}

					if newKey.Type() != f.Type().Key() {
						newKey = newKey.Convert(f.Type().Key())
					}

					if elem != nil {
						newVal = reflect.ValueOf(elem)
					} else {
						newVal = reflect.Zero(f.Type().Elem())
					}

					if newVal.Type() != f.Type().Elem() {
						switch newVal.Kind() {
						case reflect.Map, reflect.Slice, reflect.Array:
							newVal = reflect.New(f.Type().Elem())
							if err := setValue(newVal.Elem(), elem); err != nil {
								return err
							}
							newVal = reflect.Indirect(newVal)
						default:
							newVal = newVal.Convert(f.Type().Elem())
						}
					}

					if newVal.Kind() == reflect.Map && newVal.Len() == 0 && newMap.Type().Elem().Kind() == emptyStruct.Type().Kind() {
						if newMap.Type().Elem().NumField() == 0 {
							newMap.SetMapIndex(newKey, emptyStruct)
						} else {
							return errors.New("Map value type is struct{}, but data returned from database is a non-empty map[interface{}]interface{}")
						}
					} else {
						newMap.SetMapIndex(newKey, newVal)
					}
				}
				f.Set(newMap)
			}

		case reflect.Struct:
			// support time.Time
			if f.Type().PkgPath() == "time" && f.Type().Name() == "Time" {
				f.Set(reflect.ValueOf(time.Unix(0, int64(value.(int)))))
				break
			}

			valMap := value.(map[interface{}]interface{})
			// iteraste over struct fields and recursively fill them up
			typeOfT := f.Type()
			numFields := f.NumField()
			for i := 0; i < numFields; i++ {
				fld := typeOfT.Field(i)
				// skip unexported fields
				if fld.PkgPath != "" {
					continue
				}

				alias := fld.Name
				tag := strings.Trim(fld.Tag.Get(aerospikeTag), " ")
				if tag != "" {
					alias = tag
				}

				if valMap[alias] != nil {
					if err := setValue(f.FieldByName(fld.Name), valMap[alias]); err != nil {
						return err
					}
				}
			}

			// set the field
			f.Set(f)
		}
	}

	return nil
}
