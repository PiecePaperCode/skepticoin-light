package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
)

// From wikipedia: https://en.wikipedia.org/wiki/Variable-length_quantity
type vlqInt uint64

func deserialize(byteArr []byte, t interface{}) (interface{}, int) {
	values := reflect.ValueOf(&t).Elem()
	lenSlice := 0
	tmp := reflect.New(values.Elem().Type()).Elem()
	num := reflect.ValueOf(t).NumField()
	counter := 0
	if len(byteArr) < int(values.Type().Size()) {
		return t, 0
	}
	for i := 0; i < num; i++ {
		value := values.Elem().Field(i)

		if value.Type().String() == "main.vlqInt" {
			result, _, size := variantLengthQuantity(byteArr[counter:])
			tmp.Field(i).SetUint(uint64(result)) // Addressable
			values.Set(tmp)
			counter += size
			lenSlice = int(result)
			continue
		}
		switch value.Kind() {
		case reflect.Struct:
			nestedStruct, n := deserialize(byteArr[counter:], value.Interface())
			tmp.Field(i).Set(reflect.ValueOf(nestedStruct))
			values.Set(tmp)
			counter += n
			break
		case reflect.Uint8:
			tmp.Field(i).SetUint(uint64(byteArr[counter]))
			values.Set(tmp)
			counter++
			break
		case reflect.Uint16:
			tmp.Field(i).SetUint(
				uint64(binary.BigEndian.Uint16(
					byteArr[counter : counter+2],
				)),
			)
			values.Set(tmp)
			counter += 2
			break
		case reflect.Uint32:
			tmp.Field(i).SetUint(
				uint64(binary.BigEndian.Uint32(
					byteArr[counter : counter+4],
				)),
			)
			values.Set(tmp)
			counter += 4
			break
		case reflect.Uint64:
			tmp.Field(i).SetUint(
				binary.BigEndian.Uint64(
					byteArr[counter : counter+8],
				),
			)
			values.Set(tmp)
			counter += 8
			break
		case reflect.Array:
			size := int(value.Type().Size())
			for n := 0; n < size; n++ {
				tmp.Field(i).Index(n).Set(reflect.ValueOf(byteArr[counter]))
				counter++
			}
			values.Set(tmp)
			break
		case reflect.Slice:
			for n := 0; n < lenSlice; n++ {
				empty := reflect.New(reflect.TypeOf(value.Interface()).Elem()).Elem()
				element, size := deserialize(byteArr[counter:], empty.Interface())
				tmp.Field(i).Set(
					reflect.Append(tmp.Field(i),
						reflect.ValueOf(element),
					),
				)
				counter += size
			}
			values.Set(tmp)
			break
		}
	}
	return values.Interface(), counter
}
func serialize(t interface{}) []byte {
	buf := new(bytes.Buffer)
	values := reflect.ValueOf(&t).Elem()
	num := reflect.ValueOf(t).NumField()
	for i := 0; i < num; i++ {
		value := values.Elem().Field(i)

		if value.Type().String() == "main.vlqInt" {
			lenSlice := values.Elem().Field(i + 1).Len()
			result := toVariantLengthQuantity(vlqInt(lenSlice))
			buf.Write(result)
			continue
		}
		switch value.Kind() {
		case reflect.Struct:
			result := serialize(value.Interface())
			buf.Write(result)
			break
		case reflect.Uint8:
			result := byte(value.Uint())
			buf.WriteByte(result)
			break
		case reflect.Uint16:
			buf16 := make([]byte, 2)
			number := uint16(value.Uint())
			binary.BigEndian.PutUint16(buf16, number)
			buf.Write(buf16)
			break
		case reflect.Uint32:
			buf32 := make([]byte, 4)
			number := uint32(value.Uint())
			binary.BigEndian.PutUint32(buf32, number)
			buf.Write(buf32)
			break
		case reflect.Uint64:
			buf64 := make([]byte, 8)
			number := value.Uint()
			binary.BigEndian.PutUint64(buf64, number)
			buf.Write(buf64)
			break
		case reflect.Array:
			var arrayBuf []byte
			for n := 0; n < value.Len(); n++ {
				arrayBuf = append(arrayBuf, byte(value.Index(n).Uint()))
			}
			buf.Write(arrayBuf)
			break
		case reflect.Slice:
			for n := 0; n < value.Len(); n++ {
				slice := reflect.New(reflect.TypeOf(value.Interface()).Elem()).Elem()
				sliceBuf := serialize(slice.Interface())
				buf.Write(sliceBuf)
			}
		}
	}
	return buf.Bytes()
}
func variantLengthQuantity(byteArr []byte) (vlqInt, []byte, int) {
	var variant []uint32
	i := 0
	for {
		b := binary.BigEndian.Uint32([]byte{0, 0, 0, byteArr[i]})
		if b < 128 {
			variant = append(variant, b)
			break
		} else {
			b = binary.BigEndian.Uint32([]byte{0, 0, 0, byteArr[i]}) % 128
			variant = append(variant, b)
		}
		i++
	}
	var result uint32 = 0
	var p float64 = 0
	for i := len(variant) - 1; i > -1; i-- {
		result += variant[i] * uint32(math.Pow(128, p))
		p++
	}
	buf := new(bytes.Buffer)
	checkError(binary.Write(buf, binary.BigEndian, result))
	return vlqInt(result), buf.Bytes(), len(variant)
}
func toVariantLengthQuantity(number vlqInt) []byte {
	numberBuffer := new(bytes.Buffer)
	checkError(binary.Write(numberBuffer, binary.BigEndian, number))
	neededBytes := 0
	for _, b := range numberBuffer.Bytes() {
		if b != 0 {
			neededBytes += 8
		}
	}
	buf := new(bytes.Buffer)
	neededBytes = neededBytes / 7
	mod := 0
	for j := neededBytes; j > -1; j-- {
		div := int(math.Pow(128, float64(j)))
		modefier := 0
		if mod == 0 {
			modefier = int(number)
		} else {
			modefier = int(number) % mod
		}
		add := 0
		if j > 0 {
			add = 128
		}
		result := (modefier / div) + add
		buf.WriteByte(byte(result))
		mod = div
	}
	return buf.Bytes()
}
